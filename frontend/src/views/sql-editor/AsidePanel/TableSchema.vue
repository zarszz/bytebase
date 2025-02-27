<template>
  <div v-if="table.id !== UNKNOWN_ID" class="table-schema">
    <div class="table-schema--header">
      <div class="table-schema--header-title mr-1 flex items-center">
        <heroicons-outline:table class="h-4 w-4 mr-1" />
        <span class="font-semibold">{{ table.name }}</span>
      </div>
      <div
        class="table-schema--header-actions flex-1 flex justify-end space-x-2"
      >
        <div class="action-edit flex items-center">
          <NTooltip trigger="hover">
            <template #trigger>
              <NButton text @click="gotoAlterSchema">
                <heroicons-outline:pencil-alt class="w-4 h-4" />
              </NButton>
            </template>
            {{ $t("database.alter-schema") }}
          </NTooltip>
        </div>
        <div class="action-close flex items-center">
          <NTooltip trigger="hover">
            <template #trigger>
              <NButton text @click="handleClosePane">
                <heroicons-outline:x class="w-4 h-4" />
              </NButton>
            </template>
            {{ $t("sql-editor.close-pane") }}
          </NTooltip>
        </div>
      </div>
    </div>
    <div class="table-schema--meta text-gray-500 text-sm">
      <div class="pb-1">
        <span>{{ table.rowCount }} rows</span>
      </div>
      <div class="flex justify-between items-center w-full text-xs py-2">
        <div class="table-schema--content-column">
          <span>Columns</span>
        </div>
        <div class="table-schema--content-column">
          <span>Data Type</span>
        </div>
      </div>
    </div>
    <div class="table-schema--content text-sm text-gray-400 overflow-y-auto">
      <div
        v-for="(column, index) in table.columnList"
        :key="index"
        class="flex justify-between items-center w-full p-1 hover:bg-link-hover"
      >
        <div class="table-schema--content-column text-gray-600">
          <span>{{ column.name }}</span>
        </div>
        <div class="table-schema--content-column">
          <span>{{ column.type }}</span>
        </div>
      </div>
    </div>
  </div>
  <div v-else class="h-full flex justify-center items-center">
    {{ $t("sql-editor.table-schema-placeholder") }}
  </div>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { stringify } from "qs";

import { UNKNOWN_ID } from "@/types";
import { useSQLEditorStore } from "@/store";

const emit = defineEmits<{
  (e: "close-pane"): void;
}>();

const sqlEditorStore = useSQLEditorStore();
const table = computed(() => sqlEditorStore.selectedTable);

const gotoAlterSchema = () => {
  if (table.value.id === UNKNOWN_ID) {
    return;
  }

  const { database } = table.value;
  const query = {
    template: "bb.issue.database.schema.update",
    name: `[${database.name}] Alter schema`,
    project: database.project.id,
    databaseList: database.id,
    sql: `ALTER TABLE ${table.value.name}`,
  };
  const url = `/issue/new?${stringify(query)}`;
  window.open(url, "_blank");
};

const handleClosePane = () => {
  emit("close-pane");
};
</script>

<style scoped>
.table-schema {
  @apply h-full space-y-2;
}

.table-schema--header {
  @apply flex items-center p-2 border-b;
}

.table-schema--meta {
  @apply px-2 py-1;
  @apply border-b;
}

.table-schema--content {
  @apply px-2 py-1;
  height: calc(100% - 116px);
}
</style>
