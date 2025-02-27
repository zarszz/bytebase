<template>
  <div class="text-lg leading-6 font-medium text-main">
    <i18n-t keypath="repository.version-control-status">
      <template #status>
        <span class="text-success"> {{ $t("common.enabled") }} </span>
      </template>
    </i18n-t>
  </div>
  <div class="mt-2 textinfolabel">
    <i18n-t keypath="repository.version-control-description-file-path">
      <template #fullPath>
        <a class="normal-link" :href="repository.webUrl" target="_blank">{{
          repository.fullPath
        }}</a>
      </template>
      <template #fullPathTemplate>
        <span class="font-medium text-main"
          >{{ state.repositoryConfig.baseDirectory }}/{{
            state.repositoryConfig.filePathTemplate
          }}</span
        >
      </template>
    </i18n-t>
    <span>&nbsp;</span>
    <i18n-t keypath="repository.version-control-description-branch">
      <template #branch>
        <span class="font-medium text-main">
          <template v-if="state.repositoryConfig.branchFilter">
            {{ state.repositoryConfig.branchFilter }}
          </template>
          <template v-else>
            {{ $t("common.default") }}
          </template>
        </span>
      </template>
    </i18n-t>
    <template v-if="state.repositoryConfig.schemaPathTemplate">
      <span>&nbsp;</span>
      <i18n-t
        keypath="repository.version-control-description-description-schema-path"
      >
        <template #schemaPathTemplate>
          <span class="font-medium text-main">{{
            state.repositoryConfig.schemaPathTemplate
          }}</span>
        </template>
      </i18n-t>
    </template>
  </div>
  <RepositoryForm
    class="mt-4"
    :allow-edit="allowEdit"
    :vcs-type="repository.vcs.type"
    :vcs-name="repository.vcs.name"
    :repository-info="repositoryInfo"
    :repository-config="state.repositoryConfig"
    :project="project"
    :schema-change-type="state.schemaChangeType"
    @change-schema-change-type="(type) => (state.schemaChangeType = type)"
    @change-repository="$emit('change-repository')"
  />
  <div v-if="allowEdit" class="mt-4 pt-4 flex border-t justify-between">
    <BBButtonConfirm
      :style="'RESTORE'"
      :button-text="$t('repository.restore-to-ui-workflow')"
      :require-confirm="true"
      :ok-text="$t('common.restore')"
      :confirm-title="$t('repository.restore-to-ui-workflow') + '?'"
      :confirm-description="$t('repository.restore-ui-workflow-description')"
      @confirm="() => restoreToUIWorkflowType(true)"
    />
    <div>
      <button
        type="button"
        class="btn-primary ml-3 inline-flex justify-center py-2 px-4"
        :disabled="!allowUpdate"
        @click.prevent="doUpdate"
      >
        {{ $t("common.update") }}
      </button>
    </div>
  </div>
  <BBModal
    v-if="state.showSetupSQLReviewCIModal"
    class="relative overflow-hidden"
    :title="$t('repository.sql-review-ci-setup')"
    @close="state.showSetupSQLReviewCIModal = false"
  >
    <div class="space-y-4 max-w-[32rem]">
      <div class="whitespace-pre-wrap">
        {{
          $t("repository.sql-review-ci-setup-modal", {
            pr: repository.vcs.type.startsWith("GITLAB")
              ? $t("repository.merge-request")
              : $t("repository.pull-request"),
          })
        }}
      </div>

      <div class="flex justify-end pt-4 gap-x-2">
        <a
          class="btn-primary items-center space-x-2 mx-2 my-2"
          :href="state.sqlReviewCIPullRequestURL"
          target="_blank"
        >
          {{
            $t("repository.sql-review-ci-setup-pr", {
              pr: repository.vcs.type.startsWith("GITLAB")
                ? $t("repository.merge-request")
                : $t("repository.pull-request"),
            })
          }}
        </a>
      </div>
    </div>
  </BBModal>
  <BBAlert
    v-if="state.showSetupSQLReviewCIFailureModal"
    :style="'CRITICAL'"
    :ok-text="$t('common.retry')"
    :title="$t('repository.sql-review-ci-setup-failed')"
    @ok="
      () => {
        state.showSetupSQLReviewCIFailureModal = false;
        createSQLReviewCI();
      }
    "
    @cancel="
      () => {
        state.showSetupSQLReviewCIFailureModal = false;
      }
    "
  >
  </BBAlert>
  <BBModal
    v-if="state.showLoadingSQLReviewPRModal"
    class="relative overflow-hidden"
    :show-close="false"
    :esc-closable="false"
    :title="$t('repository.sql-review-ci-setup')"
  >
    <div
      class="whitespace-pre-wrap max-w-[32rem] flex justify-start items-start gap-x-2"
    >
      <BBSpin class="mt-1" />
      {{
        $t("repository.sql-review-ci-loading-modal", {
          pr: repository.vcs.type.startsWith("GITLAB")
            ? $t("repository.merge-request")
            : $t("repository.pull-request"),
        })
      }}
    </div>
  </BBModal>
  <BBModal
    v-if="state.showRestoreSQLReviewCIModal"
    class="relative overflow-hidden"
    :title="$t('repository.sql-review-ci-remove')"
    @close="onSQLReviewCIModalClose"
  >
    <div class="space-y-4 max-w-[32rem]">
      <div class="whitespace-pre-wrap">
        <i18n-t keypath="repository.sql-review-ci-restore-modal">
          <template #vcs>
            {{
              repository.vcs.type.startsWith("GITLAB")
                ? "GitLab CI"
                : "GitHub Action"
            }}
          </template>
          <template #repository>
            <a class="normal-link" :href="repository.webUrl" target="_blank">{{
              repository.fullPath
            }}</a>
          </template>
        </i18n-t>
      </div>

      <div class="flex justify-end pt-4 gap-x-2">
        <button
          type="button"
          class="btn-normal"
          @click.prevent="onSQLReviewCIModalClose"
        >
          {{ $t("common.close") }}
        </button>
      </div>
    </div>
  </BBModal>
  <BBModal
    v-if="state.showDisableSQLReviewCIModal"
    class="relative overflow-hidden"
    :title="$t('repository.sql-review-ci-remove')"
    @close="state.showDisableSQLReviewCIModal = false"
  >
    <div class="space-y-4 max-w-[32rem]">
      <div class="whitespace-pre-wrap">
        <i18n-t keypath="repository.sql-review-ci-remove-modal">
          <template #vcs>
            {{
              repository.vcs.type.startsWith("GITLAB")
                ? "GitLab CI"
                : "GitHub Action"
            }}
          </template>
          <template #repository>
            <a class="normal-link" :href="repository.webUrl" target="_blank">{{
              repository.fullPath
            }}</a>
          </template>
        </i18n-t>
      </div>

      <div class="flex justify-end pt-4 gap-x-2">
        <button
          type="button"
          class="btn-normal"
          @click.prevent="state.showDisableSQLReviewCIModal = false"
        >
          {{ $t("common.close") }}
        </button>
      </div>
    </div>
  </BBModal>
  <FeatureModal
    v-if="state.showFeatureModal"
    feature="bb.feature.vcs-sql-review"
    @cancel="state.showFeatureModal = false"
  />
</template>

<script lang="ts">
import { computed, defineComponent, PropType, reactive, watch } from "vue";
import isEmpty from "lodash-es/isEmpty";
import RepositoryForm from "./RepositoryForm.vue";
import {
  Repository,
  RepositoryPatch,
  ExternalRepositoryInfo,
  RepositoryConfig,
  Project,
  ProjectPatch,
  SchemaChangeType,
} from "../types";
import { useI18n } from "vue-i18n";
import {
  hasFeature,
  pushNotification,
  useProjectStore,
  useRepositoryStore,
} from "@/store";
import { isDev } from "@/utils";

interface LocalState {
  repositoryConfig: RepositoryConfig;
  schemaChangeType: SchemaChangeType;
  showFeatureModal: boolean;
  showSetupSQLReviewCIModal: boolean;
  showSetupSQLReviewCIFailureModal: boolean;
  showLoadingSQLReviewPRModal: boolean;
  showDisableSQLReviewCIModal: boolean;
  showRestoreSQLReviewCIModal: boolean;
  sqlReviewCIPullRequestURL: string;
}

export default defineComponent({
  name: "RepositoryPanel",
  components: { RepositoryForm },
  props: {
    project: {
      required: true,
      type: Object as PropType<Project>,
    },
    repository: {
      required: true,
      type: Object as PropType<Repository>,
    },
    allowEdit: {
      default: true,
      type: Boolean,
    },
  },
  emits: ["change-repository"],
  setup(props) {
    const { t } = useI18n();
    const repositoryStore = useRepositoryStore();
    const state = reactive<LocalState>({
      repositoryConfig: {
        baseDirectory: props.repository.baseDirectory,
        branchFilter: props.repository.branchFilter,
        filePathTemplate: props.repository.filePathTemplate,
        schemaPathTemplate: props.repository.schemaPathTemplate,
        sheetPathTemplate: props.repository.sheetPathTemplate,
        enableSQLReviewCI: props.repository.enableSQLReviewCI,
      },
      schemaChangeType: props.project.schemaChangeType,
      showFeatureModal: false,
      showSetupSQLReviewCIModal: false,
      showSetupSQLReviewCIFailureModal: false,
      showLoadingSQLReviewPRModal: false,
      showDisableSQLReviewCIModal: false,
      showRestoreSQLReviewCIModal: false,
      sqlReviewCIPullRequestURL: "",
    });

    watch(
      () => props.repository,
      (cur) => {
        state.repositoryConfig = {
          baseDirectory: cur.baseDirectory,
          branchFilter: cur.branchFilter,
          filePathTemplate: cur.filePathTemplate,
          schemaPathTemplate: cur.schemaPathTemplate,
          sheetPathTemplate: cur.sheetPathTemplate,
          enableSQLReviewCI: cur.enableSQLReviewCI,
        };
      }
    );

    const repositoryInfo = computed((): ExternalRepositoryInfo => {
      return {
        externalId: props.repository.externalId,
        name: props.repository.name,
        fullPath: props.repository.fullPath,
        webUrl: props.repository.webUrl,
      };
    });

    const allowUpdate = computed(() => {
      return (
        !isEmpty(state.repositoryConfig.filePathTemplate) &&
        (props.repository.branchFilter !==
          state.repositoryConfig.branchFilter ||
          props.repository.baseDirectory !==
            state.repositoryConfig.baseDirectory ||
          props.repository.filePathTemplate !==
            state.repositoryConfig.filePathTemplate ||
          props.repository.schemaPathTemplate !==
            state.repositoryConfig.schemaPathTemplate ||
          props.repository.sheetPathTemplate !==
            state.repositoryConfig.sheetPathTemplate ||
          props.repository.enableSQLReviewCI !==
            state.repositoryConfig.enableSQLReviewCI ||
          props.project.schemaChangeType !== state.schemaChangeType)
      );
    });

    const disableSQLReviewCI = computed(() => {
      return (
        props.repository.enableSQLReviewCI &&
        !state.repositoryConfig.enableSQLReviewCI
      );
    });

    const onSQLReviewCIModalClose = () => {
      state.showRestoreSQLReviewCIModal = false;
      restoreToUIWorkflowType(false);
    };

    const restoreToUIWorkflowType = (checkSQLReviewCI: boolean) => {
      if (checkSQLReviewCI && props.repository.enableSQLReviewCI) {
        state.showRestoreSQLReviewCIModal = true;
        return;
      }
      repositoryStore.deleteRepositoryByProjectId(props.project.id).then(() => {
        pushNotification({
          module: "bytebase",
          style: "SUCCESS",
          title: t("repository.restore-ui-workflow-success"),
        });
      });
    };

    const createSQLReviewCI = async () => {
      const repository = repositoryStore.getRepositoryByProjectId(
        props.project.id
      );

      state.showLoadingSQLReviewPRModal = true;

      try {
        const sqlReviewCISetup = await repositoryStore.createSQLReviewCI({
          projectId: props.project.id,
          repositoryId: repository.id,
        });
        state.sqlReviewCIPullRequestURL = sqlReviewCISetup.pullRequestURL;
        state.showSetupSQLReviewCIModal = true;
        window.open(sqlReviewCISetup.pullRequestURL, "_blank");
        repositoryStore.setRepositorySQLReviewCIEnabled({
          projectId: props.project.id,
          sqlReviewCIEnabled: true,
        });
      } catch {
        state.showSetupSQLReviewCIFailureModal = true;
      } finally {
        state.showLoadingSQLReviewPRModal = false;
      }
    };

    const doUpdate = async () => {
      if (
        state.repositoryConfig.enableSQLReviewCI &&
        !props.repository.enableSQLReviewCI &&
        !hasFeature("bb.feature.vcs-sql-review")
      ) {
        state.showFeatureModal = true;
        return;
      }

      const needSetupCI =
        !props.repository.enableSQLReviewCI &&
        state.repositoryConfig.enableSQLReviewCI;

      const repositoryPatch: RepositoryPatch = {};
      if (
        props.repository.branchFilter != state.repositoryConfig.branchFilter
      ) {
        repositoryPatch.branchFilter = state.repositoryConfig.branchFilter;
      }
      if (
        props.repository.baseDirectory != state.repositoryConfig.baseDirectory
      ) {
        repositoryPatch.baseDirectory = state.repositoryConfig.baseDirectory;
      }
      if (
        props.repository.filePathTemplate !=
        state.repositoryConfig.filePathTemplate
      ) {
        repositoryPatch.filePathTemplate =
          state.repositoryConfig.filePathTemplate;
      }
      if (
        props.repository.schemaPathTemplate !=
        state.repositoryConfig.schemaPathTemplate
      ) {
        repositoryPatch.schemaPathTemplate =
          state.repositoryConfig.schemaPathTemplate;
      }
      if (
        props.repository.sheetPathTemplate !=
        state.repositoryConfig.sheetPathTemplate
      ) {
        repositoryPatch.sheetPathTemplate =
          state.repositoryConfig.sheetPathTemplate;
      }
      if (
        props.repository.enableSQLReviewCI !=
        state.repositoryConfig.enableSQLReviewCI
      ) {
        repositoryPatch.enableSQLReviewCI =
          state.repositoryConfig.enableSQLReviewCI;
      }

      // Update project schemaChangeType field firstly.
      if (
        isDev() &&
        state.schemaChangeType !== props.project.schemaChangeType
      ) {
        const projectPatch: ProjectPatch = {
          schemaChangeType: state.schemaChangeType,
        };
        await useProjectStore().patchProject({
          projectId: props.project.id,
          projectPatch,
        });
      }

      const disableSQLReview = disableSQLReviewCI.value;

      await repositoryStore.updateRepositoryByProjectId({
        projectId: props.project.id,
        repositoryPatch,
      });

      if (needSetupCI) {
        createSQLReviewCI();
      } else if (disableSQLReview) {
        state.showDisableSQLReviewCIModal = true;
      }

      pushNotification({
        module: "bytebase",
        style: "SUCCESS",
        title: t("repository.update-version-control-config-success"),
      });
    };

    return {
      state,
      repositoryInfo,
      allowUpdate,
      restoreToUIWorkflowType,
      onSQLReviewCIModalClose,
      doUpdate,
      createSQLReviewCI,
    };
  },
});
</script>
