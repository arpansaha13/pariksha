<template>
  <UButton label="Submit" @click="showConfirmDialog" />
</template>

<script setup lang="ts">
const emit = defineEmits(['submit'])

const confirmModal = inject(InjectionKeys.ConfirmModal)!

async function showConfirmDialog() {
  const shouldSubmit = await confirmModal.open({
    title: 'Submit and end exam',
    description: 'Are you sure you want to submit and end the exam?',
    confirmLabel: 'Submit',
    variant: 'primary',
  })

  if (shouldSubmit) {
    emit('submit')
  }
}
</script>
