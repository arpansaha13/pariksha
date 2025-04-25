<template>
  <UModal
    :title="title"
    :description="description"
    :ui="{ footer: 'justify-end' }"
  >
    <template #footer>
      <UButton
        :label="confirmLabel"
        :color="variant"
        variant="solid"
        loading-auto
        @click="$emit('close', true)"
      />
      <UButton
        label="Cancel"
        color="neutral"
        variant="outline"
        @click="$emit('close', false)"
      />
    </template>
  </UModal>
</template>

<script setup lang="ts">
defineProps({
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    required: true,
  },
  confirmLabel: {
    type: String,
    default: 'Confirm',
  },
  variant: {
    type: String as PropType<'primary' | 'error'>,
    default: 'error',
  },
})
defineEmits<{
  close: [boolean]
}>()
</script>
