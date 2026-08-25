<script setup lang="ts">
import { computed } from 'vue'
import { BarChart3, Info, TrendingUp } from 'lucide-vue-next'
import { Separator } from '../ui'
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

const onUpdate = (val: string | number) => {
  if (props.disabled) return
  emit('update:modelValue', val as ScaleType)
}
</script>

<template>
  <div class="space-y-3" :class="{ 'opacity-60': props.disabled }">
    <p class="text-sm font-medium">Data Scale</p>
    <div class="flex items-center gap-2">
      <SelectionTabs
        class="min-w-0 flex-1"
        :model-value="value"
        :options="scaleOptions"
        :disabled="props.disabled"
        @update:model-value="onUpdate"
      />
      <button
        v-if="logInfo"
        type="button"
        data-testid="scale-log-info"
        class="inline-flex shrink-0 text-muted-foreground hover:text-foreground"
        :title="logInfo"
        :aria-label="logInfo"
      >
        <Info class="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  </div>
  <Separator />
</template>
