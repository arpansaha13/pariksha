<template>
  <div class="flex w-full max-w-sm flex-col items-center gap-6">
    <div>
      <Logo class="mx-auto size-20" />

      <h1 class="mb-2 mt-4 text-center text-2xl font-bold">
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
        :validate-on="['submit']"
        @submit.prevent="onSubmit"
      >
        <div class="space-y-4">
          <UFormGroup label="Email" name="email" required>
            <UInput
              v-model="signupFormData.email"
              type="email"
              placeholder="Enter your email"
              autocomplete="email"
              required
            />
          </UFormGroup>

          <UFormGroup label="Confirm email" name="confirmEmail" required>
            <UInput
              v-model="signupFormData.confirmEmail"
              type="email"
              placeholder="Re-enter your email"
              autocomplete="email"
              required
            />
          </UFormGroup>

          <UFormGroup label="Password" name="password" required>
            <UInput
              v-model="signupFormData.password"
              type="password"
              placeholder="Enter your password"
              autocomplete="current-password"
              required
            />
          </UFormGroup>

          <UFormGroup label="Confirm password" name="confirmPassword" required>
            <UInput
              v-model="signupFormData.confirmPassword"
              type="password"
              placeholder="Re-enter your password"
              autocomplete="current-password"
              required
            />
          </UFormGroup>

          <UButton type="submit" color="primary" block :loading="loading">
            Create account
          </UButton>
        </div>
      </UForm>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { FormError } from '#ui/types'

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

function validate(formState: Readonly<SignUpFormData>): FormError[] {
  const errors = []

  if (formState.email !== formState.confirmEmail) {
    errors.push({ path: 'confirmEmail', message: 'Emails do not match' })
  }

  if (formState.password !== formState.confirmPassword) {
    errors.push({ path: 'confirmPassword', message: 'Passwords do not match' })
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
    navigateTo('/auth/verification')
  } catch (err) {
    let message = 'Something went wrong.'

    if ((err as Response).status === 409) {
      message = 'This email is already registered.'
    }

    toast.add({
      id: ToastId.SIGNUP_FAILED,
      color: 'red',
      title: 'Failed to create account',
      description: message,
      icon: 'i-heroicons-exclamation-circle',
    })
  } finally {
    loading.value = false
  }
}
</script>
