<script setup lang="ts">
import { Check, ChevronsUpDown, Search } from "lucide-vue-next";
import { ref, computed, watch } from "vue";
import { cn } from "../lib/utils";
import {
  Combobox,
  ComboboxAnchor,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxItemIndicator,
  ComboboxList,
  ComboboxTrigger,
} from "./ui/combobox";
import type { Benchmark } from "../types/benchmark";

const props = defineProps<{
  benchmarks: Benchmark[];
  activeBenchmarkId: number;
  placeholder: string;
}>();

const emit = defineEmits<{
  select: [id: number];
}>();

// Convert the benchmarks array to the format expected by the combobox
const benchmarkOptions = computed(() =>
  props.benchmarks.map((b, index) => ({
    value: index.toString(),
    label: b.name,
  }))
);

// The current selected benchmark as an object
const value = ref<{ value: string; label: string } | undefined>();

// Control open/close state
const open = ref(false);

// Search term for filtering
const searchTerm = ref("");

// Filter function for combobox
const filterFunction = (list: typeof benchmarkOptions.value, searchValue: string) => {
  if (!searchValue) return list;
  return list.filter((item) =>
    item.label.toLowerCase().includes(searchValue.toLowerCase())
  );
};

// Initialize the value when the component mounts or when activeBenchmarkId changes
const updateValue = () => {
  const option = benchmarkOptions.value.find(
    (opt) => opt.value === props.activeBenchmarkId.toString()
  );
  if (option) {
    value.value = option;
  }
};

// Watch for changes in activeBenchmarkId
watch(
  () => props.activeBenchmarkId,
  () => {
    updateValue();
  },
  { immediate: true }
);

// Handle selection changes
watch(value, (newValue) => {
  if (newValue) {
    const index = parseInt(newValue.value);
    if (!isNaN(index) && index !== props.activeBenchmarkId) {
      emit("select", index);
    }
  }
});

// Handle open state changes
watch(open, (isOpen) => {
  // Close when selection is made
  if (!isOpen && value.value) {
    const index = parseInt(value.value.value);
    if (!isNaN(index) && index !== props.activeBenchmarkId) {
      emit("select", index);
    }
  }
});
</script>

<template>
  <Combobox
    v-if="benchmarks.length > 1"
    v-model:open="open"
    v-model="value"
    v-model:searchTerm="searchTerm"
    :filter-function="filterFunction"
    by="label"
    class="relative w-64"
  >
    <ComboboxAnchor>
      <ComboboxTrigger
        class="inline-flex h-10 w-full items-center justify-between rounded-lg border border-border bg-card px-4 py-2 text-sm font-medium text-card-foreground shadow-sm hover:bg-accent hover:text-accent-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
      >
        <span class="flex-1 text-center">{{
          value?.label ?? "Select benchmark..."
        }}</span>
        <ChevronsUpDown class="h-4 w-4 shrink-0 opacity-50" />
      </ComboboxTrigger>
    </ComboboxAnchor>

    <ComboboxList>
      <div v-if="benchmarkOptions.length > 10" class="sticky top-0 z-10 bg-popover -mx-1 -mt-1 border-b">
        <div class="relative w-full items-center">
          <ComboboxInput class="w-full pl-9" :placeholder="placeholder" />
          <span
            class="absolute start-0 inset-y-0 flex items-center justify-center px-3"
          >
            <Search class="size-4 text-muted-foreground" />
          </span>
        </div>
      </div>

      <ComboboxEmpty> No benchmark found.</ComboboxEmpty>

      <ComboboxGroup>
        <ComboboxItem
          v-for="benchmark in benchmarkOptions"
          :key="benchmark.value"
          :value="benchmark"
          :text-value="benchmark.label"
          class="flex items-center justify-between"
        >
          <span class="flex-1 text-center">{{ benchmark.label }}</span>

          <ComboboxItemIndicator>
            <Check :class="cn('h-4 w-4')" />
          </ComboboxItemIndicator>
        </ComboboxItem>
      </ComboboxGroup>
    </ComboboxList>
  </Combobox>
</template>
