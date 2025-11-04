<template>
  <UButton
    label="Submit"
    icon="heroicons:cloud-arrow-up"
    class="hidden sm:flex"
    @click="showConfirmDialog"
  />
  <UButton
    size="sm"
    icon="heroicons:cloud-arrow-up"
    class="sm:hidden"
    @click="showConfirmDialog"
  />
</template>

<script setup lang="ts">
const emit = defineEmits(['submit'])

const confirmModal = inject(InjectionKeys.ConfirmModal)!

async function showConfirmDialog() {
  const instance = confirmModal.open({
    title: 'Submit and end exam',
    description: 'Are you sure you want to submit and end the exam?',
    confirmLabel: 'Submit',
    variant: 'primary',
  })

  const shouldSubmit = await instance.result
  if (shouldSubmit) {
    emit('submit')
  }
}
</script>
