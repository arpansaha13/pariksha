<template>
  <div class="flex w-full max-w-sm flex-col items-center gap-6">
    <div>
      <Logo class="mx-auto size-20" />

      <h1 class="mb-2 mt-4 text-center text-2xl font-bold">
        <span v-if="!isResetComplete">Reset Password</span>
        <template v-else>
          <span>Your password has been reset successfully!</span>
          <Icon name="twemoji:party-popper" class="ml-[1ch]" />
        </template>
      </h1>

      <p v-if="!isResetComplete" class="text-pretty text-center text-gray-700">
        Please check your email
        <ClientOnly>
          <span class="font-medium">{{ authStore.forgotPassEmail }}</span>
        </ClientOnly>
        and enter the OTP along with your new password below.
      </p>

      <template v-else>
        <p class="mb-4 text-pretty text-center text-gray-700">
          You can now log in to your account with your new password.
        </p>

        <UButton to="/auth/login" replace class="mx-auto flex w-max">
          Go to Login
        </UButton>
      </template>
    </div>

    <UCard v-if="!isResetComplete" class="w-full">
      <UForm
        :state="resetPasswordFormData"
        :validate="validate"
        :validate-on="['submit']"
        @submit.prevent="onSubmit"
      >
        <div class="space-y-4">
          <UFormGroup label="OTP" name="otp" required>
            <UInput
              v-model="resetPasswordFormData.otp"
              type="text"
              inputmode="numeric"
              autocomplete="off"
              maxlength="6"
              required
            />
          </UFormGroup>

          <UFormGroup label="New Password" name="newPassword" required>
            <UInput
              v-model="resetPasswordFormData.newPassword"
              type="password"
              autocomplete="new-password"
              required
            />
          </UFormGroup>

          <UFormGroup
            label="Confirm New Password"
            name="confirmNewPassword"
            required
          >
            <UInput
              v-model="resetPasswordFormData.confirmNewPassword"
              type="password"
              autocomplete="new-password"
              required
            />
          </UFormGroup>

          <UButton type="submit" color="primary" block :loading="loading">
            Reset Password
          </UButton>
        </div>
      </UForm>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { FormError } from '#ui/types'

interface ResetPasswordFormData {
  otp: string
  newPassword: string
  confirmNewPassword: string
}

definePageMeta({
  layout: 'auth',
})

useHead({
  title: 'Reset Password',
})

const toast = useToast()
const authStore = useAuthStore()
const router = useRouter()

onBeforeMount(() => {
  // Redirect if email not set
  if (!authStore.forgotPassEmail) {
    router.replace('/auth/forgot-password')
  }
})

const resetPasswordFormData = useState<ResetPasswordFormData>(() => ({
  otp: '',
  newPassword: '',
  confirmNewPassword: '',
}))

const loading = useState(() => false)
const isResetComplete = useState(() => false)

function validate(formState: Readonly<ResetPasswordFormData>): FormError[] {
  const errors = []

  if (formState.newPassword !== formState.confirmNewPassword) {
    errors.push({
      path: 'confirmNewPassword',
      message: 'Passwords do not match',
    })
  }

  return errors
}

async function onSubmit() {
  try {
    loading.value = true

    await resetPassword({
      email: authStore.forgotPassEmail!,
      new_password: resetPasswordFormData.value.newPassword,
      otp: resetPasswordFormData.value.otp,
    })

    authStore.clearForgotPassEmail()
    isResetComplete.value = true
  } catch (err) {
    const status = (err as Response).status
    let message = 'Something went wrong.'

    if (status === 400) {
      message = 'Invalid OTP. Please try again.'
    } else if ([403, 404].includes(status)) {
      message = 'This email is not registered.'
    }

    toast.add({
      id: ToastId.RESET_PASSWORD_FAILED,
      color: 'red',
      title: 'Failed to reset password',
      description: message,
      icon: 'i-heroicons-exclamation-circle',
    })
  } finally {
    loading.value = false
  }
}
</script>
