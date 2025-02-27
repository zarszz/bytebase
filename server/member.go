package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/jsonapi"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/api"
	"github.com/bytebase/bytebase/common"
)

func (s *Server) registerMemberRoutes(g *echo.Group) {
	g.POST("/member", func(c echo.Context) error {
		ctx := c.Request().Context()
		memberCreate := &api.MemberCreate{}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, memberCreate); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformed create member request").SetInternal(err)
		}

		memberCreate.CreatorID = c.Get(getPrincipalIDContextKey()).(int)

		member, err := s.store.CreateMember(ctx, memberCreate)
		if err != nil {
			if common.ErrorCode(err) == common.Conflict {
				return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("Member for user ID already exists: %d", memberCreate.PrincipalID))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create member").SetInternal(err)
		}

		// Record activity
		{
			user, err := s.store.GetPrincipalByID(ctx, member.PrincipalID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Server error to find user ID: %d", member.PrincipalID)).SetInternal(err)
			}
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Failed to find user ID: %d", member.PrincipalID))
			}

			bytes, err := json.Marshal(api.ActivityMemberCreatePayload{
				PrincipalID:    member.PrincipalID,
				PrincipalName:  user.Name,
				PrincipalEmail: user.Email,
				MemberStatus:   member.Status,
				Role:           member.Role,
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to construct activity payload").SetInternal(err)
			}
			activityCreate := &api.ActivityCreate{
				CreatorID:   c.Get(getPrincipalIDContextKey()).(int),
				ContainerID: member.ID,
				Type:        api.ActivityMemberCreate,
				Level:       api.ActivityInfo,
				Payload:     string(bytes),
			}
			_, err = s.ActivityManager.CreateActivity(ctx, activityCreate, &ActivityMeta{})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create activity after creating member: %d", member.ID)).SetInternal(err)
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, member); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal create member response").SetInternal(err)
		}
		return nil
	})

	g.GET("/member", func(c echo.Context) error {
		ctx := c.Request().Context()
		memberFind := &api.MemberFind{}
		memberList, err := s.store.FindMember(ctx, memberFind)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch member list").SetInternal(err)
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, memberList); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal member list response").SetInternal(err)
		}
		return nil
	})

	g.PATCH("/member/:id", func(c echo.Context) error {
		ctx := c.Request().Context()
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		member, err := s.store.GetMemberByID(ctx, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Server error to find member ID: %d", id)).SetInternal(err)
		}
		if member == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Failed to find member ID: %d", id))
		}

		memberPatch := &api.MemberPatch{
			ID:        id,
			UpdaterID: c.Get(getPrincipalIDContextKey()).(int),
		}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, memberPatch); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformed patch member request").SetInternal(err)
		}
		// When archiving an owner, make sure there are other active owners.
		if member.Role == api.Owner && memberPatch.RowStatus != nil && *memberPatch.RowStatus == string(api.Archived) {
			countResult, err := s.store.CountMemberGroupByRoleAndStatus(ctx)
			for _, count := range countResult {
				if count.Role == api.Owner && count.RowStatus == api.Normal && count.Count == 1 {
					return echo.NewHTTPError(http.StatusInternalServerError, "Cannot archive the only remaining owner in workspace").SetInternal(err)
				}
			}
		}

		updatedMember, err := s.store.PatchMember(ctx, memberPatch)
		if err != nil {
			if common.ErrorCode(err) == common.NotFound {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Member ID not found: %d", id))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to patch member ID: %v", id)).SetInternal(err)
		}

		// Record activity
		{
			user, err := s.store.GetPrincipalByID(ctx, updatedMember.PrincipalID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Server error to find user ID: %d", updatedMember.PrincipalID)).SetInternal(err)
			}
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Failed to find user ID: %d", updatedMember.PrincipalID))
			}

			if memberPatch.Role != nil {
				bytes, err := json.Marshal(api.ActivityMemberRoleUpdatePayload{
					PrincipalID:    updatedMember.PrincipalID,
					PrincipalName:  user.Name,
					PrincipalEmail: user.Email,
					OldRole:        member.Role,
					NewRole:        updatedMember.Role,
				})
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to construct activity payload").SetInternal(err)
				}
				activityCreate := &api.ActivityCreate{
					CreatorID:   c.Get(getPrincipalIDContextKey()).(int),
					ContainerID: updatedMember.ID,
					Type:        api.ActivityMemberRoleUpdate,
					Level:       api.ActivityInfo,
					Payload:     string(bytes),
				}
				_, err = s.ActivityManager.CreateActivity(ctx, activityCreate, &ActivityMeta{})
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create activity after changing member role: %d", updatedMember.ID)).SetInternal(err)
				}
			} else if memberPatch.RowStatus != nil {
				bytes, err := json.Marshal(api.ActivityMemberActivateDeactivatePayload{
					PrincipalID:    updatedMember.PrincipalID,
					PrincipalName:  user.Name,
					PrincipalEmail: user.Email,
					Role:           member.Role,
				})
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to construct activity payload").SetInternal(err)
				}
				theType := api.ActivityMemberActivate
				if *memberPatch.RowStatus == "ARCHIVED" {
					theType = api.ActivityMemberDeactivate
				}
				activityCreate := &api.ActivityCreate{
					CreatorID:   c.Get(getPrincipalIDContextKey()).(int),
					ContainerID: updatedMember.ID,
					Type:        theType,
					Level:       api.ActivityInfo,
					Payload:     string(bytes),
				}
				_, err = s.ActivityManager.CreateActivity(ctx, activityCreate, &ActivityMeta{})
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create activity after changing member role: %d", updatedMember.ID)).SetInternal(err)
				}
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, updatedMember); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal member ID response: %v", id)).SetInternal(err)
		}
		return nil
	})
}

// getAnyFromWorkspaceOwnerOrDBA finds a default assignee from the workspace owners or DBAs.
func (s *Server) getAnyWorkspaceOwnerOrDBA(ctx context.Context) (*api.Member, error) {
	for _, role := range []api.Role{api.Owner, api.DBA} {
		memberList, err := s.store.FindMember(ctx, &api.MemberFind{
			Role: &role,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get role %v", role)
		}
		if len(memberList) > 0 {
			return memberList[0], nil
		}
	}
	return nil, errors.New("failed to get a workspace owner or DBA")
}
