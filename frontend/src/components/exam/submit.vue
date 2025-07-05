<template>
  <UButton
    label="Submit"
    icon="heroicons:cloud-arrow-up"
    @click="showConfirmDialog"
    class="hidden sm:flex"
  />
  <UButton
    size="sm"
    icon="heroicons:cloud-arrow-up"
    @click="showConfirmDialog"
    class="sm:hidden"
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
