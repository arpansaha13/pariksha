<template>
  <UModal
    :title="title"
    :description="description"
    :ui="{ footer: 'justify-end' }"
  >
    <template #footer>
      <UButton
        label="Confirm deletion"
        color="error"
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
})
defineEmits<{
  close: [boolean]
}>()
</script>
