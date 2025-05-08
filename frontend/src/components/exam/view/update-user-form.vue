<template>
  <UCard>
    <template #header>
      <h2 class="heading">Verify your information</h2>
    </template>

    <UForm
      :state="formState"
      :validate="validate"
      :validate-on="['blur']"
      class="flex flex-col gap-y-4"
      @submit="onSubmit"
    >
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
import type { FormError } from '@nuxt/ui'

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

function validate(formState: ExamViewUpdateUserFormState): FormError[] {
  const errors = []

  if (!isAlpha(formState.first_name)) {
    errors.push({
      name: 'first_name',
      message: 'Name can only have alphabets',
    })
  }

  if (!isAlpha(formState.last_name)) {
    errors.push({
      name: 'last_name',
      message: 'Name can only have alphabets',
    })
  }

  return errors
}

async function onSubmit() {
  emit('submit', formState.value)
}
</script>
