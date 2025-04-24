<template>
  <div class="flex w-full max-w-sm flex-col items-center gap-6">
    <div>
      <Logo class="mx-auto size-20" />

      <h1 class="mt-4 mb-2 text-center text-2xl font-bold">Forgot Password?</h1>

      <p class="mb-5 text-center text-pretty text-gray-700">
        No worries! Just enter your email address, and we'll promptly send an
        OTP to your inbox to help you reset your password.
      </p>

      <p class="text-center">
        <span>Remember your password?</span>
        <ULink
          to="/auth/login"
          class="text-primary-500 hover:text-primary-600 font-semibold"
        >
          Login<span class="font-normal text-black">.</span>
        </ULink>
      </p>
    </div>

    <UCard class="w-full">
      <UForm :state="forgotPasswordFormData" @submit.prevent="onSubmit">
        <div class="space-y-4">
          <UFormField label="Email" name="email">
            <UInput
              v-model="forgotPasswordFormData.email"
              required
              type="email"
              autocomplete="email"
              placeholder="Enter your email"
              class="w-full"
            />
          </UFormField>

          <UButton type="submit" color="primary" block :loading="loading">
            Send OTP
          </UButton>
        </div>
      </UForm>
    </UCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'auth',
})

useHead({
  title: 'Forgot Password',
})

const toast = useToast()
const authStore = useAuthStore()

const forgotPasswordFormData = useState(() => ({
  email: '',
}))

const loading = useState(() => false)

async function onSubmit() {
  try {
    loading.value = true
    await forgotPassword(forgotPasswordFormData.value)

    // Store email for reset password page
    authStore.setForgotPassEmail(forgotPasswordFormData.value.email)

    await navigateTo('/auth/reset-password')
  } catch (err) {
    const status = (err as Response).status
    let message = 'Something went wrong.'

    if (status === 404) {
      message = 'This email is not registered.'
    }

    toast.clear()
    toast.add({
      id: ToastId.FORGOT_PASSWORD_FAILED,
      color: 'error',
      title: 'Failed to send OTP',
      description: message,
      icon: 'i-heroicons-exclamation-circle',
    })
  } finally {
    loading.value = false
  }
}
</script>
