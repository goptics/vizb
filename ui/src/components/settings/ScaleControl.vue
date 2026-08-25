<script setup lang="ts">
import { computed, ref } from 'vue'
import { BarChart3, Info, TrendingUp } from 'lucide-vue-next'
import { Popover, PopoverContent, PopoverTrigger, Separator } from '../ui'
import SelectionTabs from '../SelectionTabs.vue'
import type { ScaleAxis, ScaleInput, ScaleType } from '@/types'
import { scaleLogInfoText } from '@/lib/scale'

const props = withDefaults(
  defineProps<{
    modelValue: ScaleInput | undefined
    disabled?: boolean
    defaultAxes?: readonly ScaleAxis[]
  }>(),
  {
    defaultAxes: () => ['y'],
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: ScaleType): void
}>()

const logInfo = computed(() => scaleLogInfoText(props.modelValue, props.defaultAxes))
const value = computed<ScaleType>(() => (logInfo.value ? 'log' : 'linear'))

const scaleOptions = [
  { value: 'linear', label: 'Linear', icon: BarChart3 },
  { value: 'log', label: 'Logarithmic', icon: TrendingUp },
]

const infoOpen = ref(false)

const onUpdate = (val: string | number) => {
  if (props.disabled) return
  emit('update:modelValue', val as ScaleType)
}

const setInfoOpen = (open: boolean) => {
  infoOpen.value = open
}
</script>

<template>
  <div class="space-y-3" :class="{ 'opacity-60': props.disabled }">
    <p class="text-sm font-medium">Data Scale</p>
    <SelectionTabs
      class="min-w-0 w-full"
      :model-value="value"
      :options="scaleOptions"
      :disabled="props.disabled"
      @update:model-value="onUpdate"
    >
      <template #suffix="{ option }">
        <span
          v-if="option.value === 'log' && logInfo"
          data-testid="scale-log-info"
          class="ml-1 inline-flex text-muted-foreground hover:text-foreground"
          :aria-label="logInfo"
          :aria-expanded="infoOpen"
          role="button"
          tabindex="0"
          @mouseenter="setInfoOpen(true)"
          @mouseleave="setInfoOpen(false)"
          @focus="setInfoOpen(true)"
          @blur="setInfoOpen(false)"
          @click.stop
          @pointerdown.stop
        >
          <Popover :open="infoOpen">
            <PopoverTrigger as-child>
              <span class="inline-flex">
                <Info class="h-4 w-4" aria-hidden="true" />
              </span>
            </PopoverTrigger>
            <PopoverContent class="w-auto max-w-xs p-3 text-sm" align="end">
              {{ logInfo }}
            </PopoverContent>
          </Popover>
        </span>
      </template>
    </SelectionTabs>
  </div>
  <Separator />
</template>
