<template>
  <component :is="as">
    <div v-if="$slots.header" class="mb-1 font-semibold">
      <slot name="header" />
    </div>

    <UCard
      variant="subtle"
      :ui="{
        root: bodyBgColor,
        body: [
          'font-mono',
          bodyPadding,
          preserveWhiteSpace && 'whitespace-pre-wrap',
        ],
      }"
    >
      <slot />
    </UCard>
  </component>
</template>

<script setup lang="ts">
const {
  as = 'div',
  color = 'neutral',
  padding = 'default',
  preserveWhiteSpace = false,
} = defineProps<{
  as?: keyof HTMLElementTagNameMap
  color?: 'neutral' | 'error'
  padding?: 'small' | 'default'
  preserveWhiteSpace?: boolean
}>()

const bodyBgColor = computed(() => {
  if (color === 'error') return 'bg-error-50 ring-error-200'
  return '' // default UCard background
})

const bodyPadding = computed(() => {
  if (padding === 'small') return '!px-3 !py-2.5'
  return '' // default UCard background
})
</script>
