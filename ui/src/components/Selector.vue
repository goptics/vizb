<script setup lang="ts">
import { ChevronsUpDown, Search } from 'lucide-vue-next'
import { ref, computed, watch, type Component, type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'
import { filterPickerOptions, limitPickerOptions } from '@/lib/pickerLimit'

import {
  Combobox,
  ComboboxAnchor,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
  ComboboxTrigger,
} from './ui/combobox'

// Selector is a generic combobox used by the dataset picker (DataSetHeader),
// stats view switcher (StatsPanel), axis swap (SwapControl), and the chart-
// type picker (SettingsPanel). The `icon` field is optional — call sites that
// don't have a meaningful icon (e.g. group names) just don't pass one and the
// trigger / item render label-only, exactly as before.
type Item = {
  name: string
  icon?: Component
  value?: string
}

const props = defineProps<{
  items: Item[]
  activeId?: number
  activeValue?: string
  placeholder?: string
  notFoundText?: string
  ariaLabel?: string
  resultLimit?: number
  class?: HTMLAttributes['class']
  triggerClass?: HTMLAttributes['class']
}>()

const emit = defineEmits<{
  select: [id: number]
  selectValue: [value: string]
}>()

// Convert the input array to the format expected by the combobox. Carries the
// `icon` through verbatim so the trigger and item can render it next to the
// label without a second pass over `items`.
const options = computed(() =>
  props.items.map((item, index) => ({
    value: item.value ?? index.toString(),
    label: item.name,
    icon: item.icon,
  }))
)

type Option = { value: string; label: string; icon?: Component }
const value = ref<Option | undefined>()

// Control open/close state
const open = ref(false)

// Search term for filtering
const searchTerm = ref('')

// Filter function for combobox. Limited selectors filter the complete option
// list here and instantiate only the capped result set below.
const filterFunction = (list: Option[], searchValue: string) => {
  if (!searchValue) return list
  return list.filter((item) => item.label.toLowerCase().includes(searchValue.toLowerCase()))
}

const matchingOptions = computed(() => filterPickerOptions(options.value, searchTerm.value))
const activeOption = computed(() => {
  const activeValue = props.activeValue ?? props.activeId?.toString()
  return options.value.find((option) => option.value === activeValue)
})
const visibleOptions = computed(() =>
  props.resultLimit
    ? limitPickerOptions(matchingOptions.value, activeOption.value, props.resultLimit)
    : options.value
)
const matchingValues = computed(() => new Set(matchingOptions.value.map((option) => option.value)))
const visibleMatchCount = computed(
  () => visibleOptions.value.filter((option) => matchingValues.value.has(option.value)).length
)
const resultsTruncated = computed(
  () => !!props.resultLimit && visibleMatchCount.value < matchingOptions.value.length
)
const comboboxFilter = (list: Option[], searchValue: string) =>
  props.resultLimit ? list : filterFunction(list, searchValue)

// Initialize the value when the component mounts or when activeId changes
const updateValue = () => {
  const activeValue = props.activeValue ?? props.activeId?.toString()
  const option = options.value.find((opt) => opt.value === activeValue)
  if (option) {
    value.value = option
  } else {
    value.value = undefined
  }
}

// Watch for changes in activeId or benchmarks list
watch(
  [() => props.activeId, () => props.activeValue, () => props.items],
  () => {
    updateValue()
  },
  { immediate: true }
)

// Handle selection changes
watch(value, (newValue) => {
  if (newValue) {
    if (props.activeValue !== undefined) {
      if (newValue.value !== props.activeValue) emit('selectValue', newValue.value)
      return
    }
    const index = parseInt(newValue.value)
    if (!isNaN(index) && index !== props.activeId) {
      emit('select', index)
    }
  }
})

// Handle open state changes
watch(open, (isOpen) => {
  // Close when selection is made
  if (!isOpen && value.value) {
    if (props.activeValue !== undefined) {
      if (value.value.value !== props.activeValue) emit('selectValue', value.value.value)
      return
    }
    const index = parseInt(value.value.value)
    if (!isNaN(index) && index !== props.activeId) {
      emit('select', index)
    }
  }
})
</script>

<template>
  <Combobox
    v-if="items.length > 1"
    v-model:open="open"
    v-model="value"
    v-model:searchTerm="searchTerm"
    :filter-function="comboboxFilter"
    by="label"
    :class="cn('relative w-64', props.class)"
  >
    <ComboboxAnchor>
      <ComboboxTrigger
        :aria-label="ariaLabel"
        :class="
          cn(
            'inline-flex h-10 w-full items-center rounded-lg border border-border bg-card px-4 text-sm font-medium text-card-foreground shadow-sm hover:bg-accent hover:text-accent-foreground',
            triggerClass
          )
        "
      >
        <span :class="value?.icon ? 'flex flex-1 items-center gap-2' : 'flex-1 text-center'">
          <component :is="value.icon" v-if="value?.icon" class="h-4 w-4" />
          <span>{{ value?.label }}</span>
        </span>
        <ChevronsUpDown class="h-4 w-4" />
      </ComboboxTrigger>
    </ComboboxAnchor>

    <ComboboxList>
      <div v-if="options.length > 10" class="sticky top-0 z-10 -mx-1 -mt-1 border-b bg-popover">
        <div class="relative items-center">
          <ComboboxInput class="pl-9" :placeholder="placeholder" />
          <span class="absolute inset-y-0 flex items-center px-3">
            <Search class="size-4 text-muted-foreground" />
          </span>
        </div>
      </div>

      <ComboboxEmpty>{{ notFoundText }}</ComboboxEmpty>

      <ComboboxGroup>
        <ComboboxItem
          v-for="option in visibleOptions"
          :key="option.value"
          :value="option"
          :text-value="option.label"
          class="flex items-center gap-2 data-[state=checked]:font-medium data-[state=checked]:text-primary"
        >
          <template v-if="option.icon">
            <component :is="option.icon" class="h-4 w-4" />
            <span>{{ option.label }}</span>
          </template>
          <span v-else class="flex-1 text-center">{{ option.label }}</span>
        </ComboboxItem>
      </ComboboxGroup>

      <p
        v-if="resultsTruncated"
        class="border-t px-2 py-2 text-center text-xs text-muted-foreground"
      >
        Showing {{ visibleMatchCount }} of {{ matchingOptions.length }} matches. Narrow your search.
      </p>
    </ComboboxList>
  </Combobox>
</template>
