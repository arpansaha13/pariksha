<template>
  <div>
    <div v-if="$slots.header" class="mb-1 font-semibold">
      <slot name="header" />
    </div>

    <UCard
      variant="subtle"
      :ui="{
        root: bodyBgColor,
        body: ['font-mono', preserveWhiteSpace && 'whitespace-pre-wrap'],
      }"
    >
      <slot />
    </UCard>
  </div>
</template>

<script setup lang="ts">
const { color = 'neutral', preserveWhiteSpace = false } = defineProps<{
  color?: 'neutral' | 'error'
  preserveWhiteSpace?: boolean
}>()

const bodyBgColor = computed(() => {
  if (color === 'error') return 'bg-error-50 ring-error-200'
  return '' // default UCard background
})
</script>
