<template>
  <UCard>
    <template #header>
      <h2 class="heading">Verify your information</h2>
    </template>

    <UForm :state="formState" class="flex flex-col gap-y-4" @submit="onSubmit">
      <UFormField label="First name" name="first_name" required>
        <UInput v-model="formState.first_name" required class="w-52" />
      </UFormField>

      <UFormField label="Last name" name="last_name" required>
        <UInput v-model="formState.last_name" required class="w-52" />
      </UFormField>

      <button ref="submitButton" type="submit" class="hidden" />
    </UForm>
  </UCard>
</template>

<script lang="ts" setup>
interface ExamViewUpdateUserFormState {
  first_name: string
  last_name: string
}

const formState = defineModel<ExamViewUpdateUserFormState>('form-data', {
  required: true,
})

const submitButtonRef = useTemplateRef('submitButton')
defineExpose({
  submit: () => submitButtonRef.value?.click(),
})

const emit = defineEmits<{
  submit: [form: ExamViewUpdateUserFormState]
}>()

async function onSubmit() {
  emit('submit', formState.value)
}
</script>

<style></style>
