<script setup lang="ts">
import { Check, ChevronsUpDown, Search } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import type { DescriptiveStats } from '../types'
import {
  DESCRIPTIVE_COLUMNS,
  STAT_CATEGORY_LABELS,
  STAT_CATEGORY_ORDER,
  columnsFromKeys,
  isDescriptiveColumnKey,
} from '../lib/descriptiveColumns'
import {
  Combobox,
  ComboboxAnchor,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxItemIndicator,
  ComboboxLabel,
  ComboboxList,
  ComboboxSeparator,
  ComboboxTrigger,
} from './ui/combobox'

const props = defineProps<{
  defaultKeys: (keyof DescriptiveStats)[]
}>()

const modelValue = defineModel<(keyof DescriptiveStats)[]>({ required: true })

const open = ref(false)
const searchTerm = ref('')

const allKeys = DESCRIPTIVE_COLUMNS.map((col) => col.key)

function normalizeKeys(keys: readonly string[]): (keyof DescriptiveStats)[] {
  const selected = new Set(keys.filter(isDescriptiveColumnKey))
  return allKeys.filter((key) => selected.has(key))
}

const selectedKeys = computed<(keyof DescriptiveStats)[]>({
  get: () => modelValue.value,
  set: (keys) => {
    const normalized = normalizeKeys(keys)
    if (normalized.length > 0) modelValue.value = normalized
  },
})

const selectedSet = computed(() => new Set(selectedKeys.value))

const groups = computed(() =>
  STAT_CATEGORY_ORDER.map((category) => ({
    category,
    label: STAT_CATEGORY_LABELS[category],
    columns: DESCRIPTIVE_COLUMNS.filter((col) => col.category === category),
  }))
)

const triggerLabel = computed(() => {
  const columns = columnsFromKeys(selectedKeys.value)
  if (columns.length === DESCRIPTIVE_COLUMNS.length) return 'All columns'
  if (columns.length <= 2) return columns.map((col) => col.label).join(', ')
  return `${columns.length} columns`
})

function filterColumns(keys: string[], query: string) {
  const q = query.trim().toLowerCase()
  if (!q) return keys
  return keys.filter((key) => {
    if (!isDescriptiveColumnKey(key)) return false
    const col = DESCRIPTIVE_COLUMNS.find((column) => column.key === key)
    if (!col) return false
    return (
      col.label.toLowerCase().includes(q) ||
      key.toLowerCase().includes(q) ||
      STAT_CATEGORY_LABELS[col.category].toLowerCase().includes(q)
    )
  })
}

function isLastSelected(key: keyof DescriptiveStats) {
  return selectedKeys.value.length === 1 && selectedSet.value.has(key)
}

function selectAll() {
  modelValue.value = allKeys
}

function resetToDefaults() {
  const defaults = normalizeKeys(props.defaultKeys)
  modelValue.value = defaults.length ? defaults : allKeys
}
</script>

<template>
  <Combobox
    v-model="selectedKeys"
    v-model:open="open"
    v-model:searchTerm="searchTerm"
    multiple
    :reset-search-term-on-select="false"
    :filter-function="filterColumns"
    class="relative w-56"
  >
    <ComboboxAnchor>
      <ComboboxTrigger
        class="inline-flex h-10 w-full items-center rounded-lg border border-border bg-card px-4 text-sm font-medium text-card-foreground shadow-sm hover:bg-accent hover:text-accent-foreground"
      >
        <span class="flex-1 truncate text-left">{{ triggerLabel }}</span>
        <ChevronsUpDown class="h-4 w-4 shrink-0" />
      </ComboboxTrigger>
    </ComboboxAnchor>

    <ComboboxList class="w-72">
      <div class="sticky top-0 z-10 -mx-1 -mt-1 border-b bg-popover">
        <div class="relative items-center">
          <ComboboxInput class="pl-9" placeholder="Search columns..." />
          <span class="absolute inset-y-0 flex items-center px-3">
            <Search class="size-4 text-muted-foreground" />
          </span>
        </div>
      </div>

      <ComboboxEmpty>No columns found.</ComboboxEmpty>

      <template v-for="(group, index) in groups" :key="group.category">
        <ComboboxSeparator v-if="index > 0" />
        <ComboboxGroup>
          <ComboboxLabel>{{ group.label }}</ComboboxLabel>
          <ComboboxItem
            v-for="column in group.columns"
            :key="column.key"
            :value="column.key"
            :text-value="column.label"
            :disabled="isLastSelected(column.key)"
            class="gap-2 data-[state=checked]:font-medium data-[state=checked]:text-primary"
          >
            <span>{{ column.label }}</span>
            <ComboboxItemIndicator>
              <Check class="h-4 w-4" />
            </ComboboxItemIndicator>
          </ComboboxItem>
        </ComboboxGroup>
      </template>

      <ComboboxSeparator />
      <div class="flex items-center justify-between gap-2 px-1 pt-1">
        <button
          type="button"
          class="rounded-sm px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          @click="selectAll"
        >
          Select all
        </button>
        <button
          type="button"
          class="rounded-sm px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          @click="resetToDefaults"
        >
          Reset defaults
        </button>
      </div>
    </ComboboxList>
  </Combobox>
</template>
