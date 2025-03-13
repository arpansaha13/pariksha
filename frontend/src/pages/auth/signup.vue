<template>
  <div class="flex w-full max-w-sm flex-col items-center gap-6">
    <div>
      <Logo class="mx-auto size-20" />

      <h1 class="mt-4 mb-2 text-center text-2xl font-bold">
        Create your account
      </h1>

      <p class="text-center">
        <span>Already have an account?</span>
        <ULink
          to="/auth/login"
          class="text-primary-500 hover:text-primary-600 font-semibold"
        >
          Login<span class="font-normal text-black">.</span>
        </ULink>
      </p>
    </div>

    <UCard class="w-full">
      <UForm
        :state="signupFormData"
        :validate="validate"
        :validate-on="['blur']"
        @submit.prevent="onSubmit"
      >
        <div class="space-y-4">
          <UFormField label="Email" name="email" required>
            <UInput
              v-model="signupFormData.email"
              required
              type="email"
              autocomplete="email"
              placeholder="Enter your email"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Confirm email" name="confirmEmail" required>
            <UInput
              v-model="signupFormData.confirmEmail"
              required
              type="email"
              autocomplete="email"
              placeholder="Re-enter your email"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Password" name="password" required>
            <UInput
              v-model="signupFormData.password"
              required
              type="password"
              autocomplete="current-password"
              placeholder="Enter your password"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Confirm password" name="confirmPassword" required>
            <UInput
              v-model="signupFormData.confirmPassword"
              required
              type="password"
              autocomplete="current-password"
              placeholder="Re-enter your password"
              class="w-full"
            />
          </UFormField>

          <UButton type="submit" color="primary" block :loading="loading">
            Create account
          </UButton>
        </div>
      </UForm>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { FormError } from '@nuxt/ui'

interface SignUpFormData {
  email: string
  confirmEmail: string
  password: string
  confirmPassword: string
}

definePageMeta({
  layout: 'auth',
})

useHead({
  title: 'Sign Up',
})

const toast = useToast()
const authStore = useAuthStore()

const signupFormData = useState<SignUpFormData>(() => ({
  email: '',
  confirmEmail: '',
  password: '',
  confirmPassword: '',
}))

const loading = useState(() => false)

function validate(formState: Partial<SignUpFormData>): FormError[] {
  const errors = []

  if (formState.email !== formState.confirmEmail) {
    errors.push({ name: 'confirmEmail', message: 'Emails do not match' })
  }

  if (formState.password !== formState.confirmPassword) {
    errors.push({ name: 'confirmPassword', message: 'Passwords do not match' })
  }

  return errors
}

async function onSubmit() {
  try {
    loading.value = true

    await signUp({
      email: signupFormData.value.email,
      password: signupFormData.value.password,
    })

    authStore.setSignUpEmail(signupFormData.value.email)
    await navigateTo('/auth/verification')
  } catch (err) {
    let message = 'Something went wrong.'

    if ((err as Response).status === 409) {
      message = 'This email is already registered.'
    }

    toast.add({
      id: ToastId.SIGNUP_FAILED,
      color: 'error',
      title: 'Failed to create account',
      description: message,
      icon: 'i-heroicons-exclamation-circle',
    })
  } finally {
    loading.value = false
  }
}
</script>
