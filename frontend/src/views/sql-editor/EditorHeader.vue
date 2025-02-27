<template>
  <div class="flex items-center justify-between h-16">
    <div class="flex items-center">
      <div class="flex-shrink-0 w-44">
        <router-link to="/" class="select-none" active-class exact-active-class>
          <img
            class="h-12 w-auto"
            src="../../assets/logo-full.svg"
            alt="Bytebase"
          />
        </router-link>
      </div>
      <div class="hidden sm:block">
        <div class="ml-4 flex items-baseline space-x-1">
          <router-link
            to="/sql-editor"
            class="router-link"
            exact-active-class="anchor-link"
            >{{ $t("sql-editor.self") }}</router-link
          >
          <router-link
            to="/sheets/my"
            class="router-link"
            exact-active-class="anchor-link"
            >{{ $t("sheet.self") }}</router-link
          >
        </div>
      </div>
    </div>
    <div>
      <div class="flex items-center space-x-3">
        <router-link to="/inbox" exact-active-class>
          <span
            v-if="inboxSummary.hasUnread"
            class="absolute rounded-full ml-4 -mt-1 h-2.5 w-2.5 bg-accent opacity-75"
          ></span>
          <heroicons-outline:bell class="w-6 h-6" />
        </router-link>
        <div class="ml-2">
          <ProfileDropdown />
        </div>
        <div class="ml-2 -mr-2 flex sm:hidden">
          <!-- Mobile menu button -->
          <button
            class="icon-link inline-flex items-center justify-center rounded-md"
            @click.prevent="state.showMobileMenu = !state.showMobileMenu"
          >
            <span class="sr-only">Open main menu</span>
            <!--
              Heroicon name: menu

              Menu open: "hidden", Menu closed: "block"
            -->
            <heroicons-solid:dots-vertical class="w-6 h-6" />
          </button>
        </div>
      </div>
    </div>
  </div>

  <!--
      Mobile menu, toggle classes based on menu state.

      Open: "block", closed: "hidden"
  -->
  <div v-if="state.showMobileMenu" class="block md:hidden">
    <router-link to="/project" class="bar-link rounded-md block px-3 py-2">
      {{ $t("common.projects") }}
    </router-link>

    <router-link to="/db" class="bar-link rounded-md block px-3 py-2">
      {{ $t("common.databases") }}
    </router-link>

    <router-link
      v-if="showInstanceItem"
      to="/instance"
      class="bar-link rounded-md block px-3 py-2"
      >{{ $t("common.instances") }}</router-link
    >

    <router-link
      to="/environment"
      class="bar-link rounded-md block px-3 py-2"
      >{{ $t("common.environments") }}</router-link
    >

    <router-link
      to="/setting/member"
      class="bar-link rounded-md block px-3 py-2"
      >{{ $t("common.settings") }}</router-link
    >
  </div>
</template>

<script lang="ts">
import { defineAction, useRegisterActions } from "@bytebase/vue-kbar";
import { computed, reactive, watchEffect, defineComponent } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useLocalStorage } from "@vueuse/core";
import ProfileDropdown from "@/components/ProfileDropdown.vue";
import { UNKNOWN_ID } from "@/types";
import { hasWorkspacePermission } from "@/utils";
import { useCurrentUser, useInboxStore } from "@/store";

interface LocalState {
  showMobileMenu: boolean;
}

export default defineComponent({
  name: "EditorHeader",
  components: { ProfileDropdown },
  setup() {
    const { t, availableLocales, locale } = useI18n();
    const inboxStore = useInboxStore();
    const router = useRouter();

    const state = reactive<LocalState>({
      showMobileMenu: false,
    });

    const currentUser = useCurrentUser();

    const showInstanceItem = computed((): boolean => {
      return hasWorkspacePermission(
        "bb.permission.workspace.manage-instance",
        currentUser.value.role
      );
    });

    const prepareInboxSummary = () => {
      // It will also be called when user logout
      if (currentUser.value.id != UNKNOWN_ID) {
        inboxStore.fetchInboxSummaryByUser(currentUser.value.id);
      }
    };

    watchEffect(prepareInboxSummary);

    const inboxSummary = computed(() => {
      return inboxStore.getInboxSummaryByUser(currentUser.value.id);
    });

    const kbarActions = computed(() => [
      defineAction({
        id: "bb.navigation.projects",
        name: "Projects",
        shortcut: ["g", "p"],
        section: t("kbar.navigation"),
        keywords: "navigation",
        perform: () => router.push({ name: "workspace.project" }),
      }),
      defineAction({
        id: "bb.navigation.databases",
        name: "Databases",
        shortcut: ["g", "d"],
        section: t("kbar.navigation"),
        keywords: "navigation db",
        perform: () => router.push({ name: "workspace.database" }),
      }),
      defineAction({
        id: "bb.navigation.instances",
        name: "Instances",
        shortcut: ["g", "i"],
        section: t("kbar.navigation"),
        keywords: "navigation",
        perform: () => router.push({ name: "workspace.instance" }),
      }),
      defineAction({
        id: "bb.navigation.environments",
        name: "Environments",
        shortcut: ["g", "e"],
        section: t("kbar.navigation"),
        keywords: "navigation",
        perform: () => router.push({ name: "workspace.environment" }),
      }),
      defineAction({
        id: "bb.navigation.settings",
        name: "Settings",
        shortcut: ["g", "s"],
        section: t("kbar.navigation"),
        keywords: "navigation",
        perform: () => router.push({ name: "setting.workspace.member" }),
      }),
      defineAction({
        id: "bb.navigation.inbox",
        name: "Inbox",
        shortcut: ["g", "m"],
        section: t("kbar.navigation"),
        keywords: "navigation",
        perform: () => router.push({ name: "setting.inbox" }),
      }),
    ]);
    useRegisterActions(kbarActions);

    const storage = useLocalStorage("bytebase_options", {}) as any;

    const setLocale = (lang: string) => {
      locale.value = lang;
      storage.value = {
        appearance: {
          language: lang,
        },
      };
    };

    const I18N_CHANGE_ACTION_ID_NAMESPACE = "bb.preferences.locale";
    const i18nChangeAction = computed(() =>
      defineAction({
        // here `id` is "bb.preferences.locale"
        id: I18N_CHANGE_ACTION_ID_NAMESPACE,
        section: t("kbar.preferences.common"),
        name: t("kbar.preferences.change-language"),
        keywords: "language lang locale",
      })
    );
    const i18nActions = computed(() => [
      i18nChangeAction.value,
      ...availableLocales.map((lang) => {
        return defineAction({
          // here `id` looks like "bb.preferences.locale.en"
          id: `${I18N_CHANGE_ACTION_ID_NAMESPACE}.${lang}`,
          name: lang,
          parent: I18N_CHANGE_ACTION_ID_NAMESPACE,
          keywords: `language lang locale ${lang}`,
          perform: () => setLocale(lang),
        });
      }),
    ]);
    useRegisterActions(i18nActions);

    const goBack = () => {
      if (window.history.state?.back) {
        router.go(-1);
      } else {
        router.push("/");
      }
    };

    return {
      state,
      showInstanceItem,
      inboxSummary,
      goBack,
    };
  },
});
</script>

<style scoped>
.router-link {
  @apply text-base ml-2 truncate px-2 py-2 rounded-lg no-underline hover:bg-gray-200;
}

.router-link-active {
  @apply bg-gray-200;
}
</style>
