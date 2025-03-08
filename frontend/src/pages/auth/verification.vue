<template>
  <div class="flex w-full max-w-sm flex-col items-center gap-6">
    <div>
      <Logo class="mx-auto size-20" />

      <h1 class="mb-2 mt-4 text-center text-2xl font-bold">
        Verify your email
      </h1>

      <p class="text-pretty text-center">
        Check your email
        <ClientOnly>
          <strong>{{ authStore.signUpEmail }}</strong>
        </ClientOnly>
        for an OTP we just sent. Once you've found it, please enter it here to
        proceed.
      </p>
    </div>

    <UCard class="w-full">
      <UForm :state="verificationFormData" @submit.prevent="onSubmit">
        <div class="space-y-4">
          <UFormGroup label="OTP" name="otp">
            <UInput
              v-model="verificationFormData.otp"
              type="text"
              inputmode="numeric"
              autocomplete="off"
              maxlength="6"
              required
            />
          </UFormGroup>

          <UButton type="submit" color="primary" block :loading="loading">
            Submit
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
  title: 'Verification',
})

const toast = useToast()
const authStore = useAuthStore()
const router = useRouter()

onBeforeMount(() => {
  // Redirect if email not set
  if (!authStore.signUpEmail) {
    router.replace('/auth/signup')
  }
})

const verificationFormData = useState(() => ({
  email: '',
  otp: '',
}))

onMounted(() => {
  verificationFormData.value.email = authStore.signUpEmail!
})

const loading = useState(() => false)

async function onSubmit() {
  try {
    loading.value = true
    await verifySignUpEmail(verificationFormData.value)
    await navigateTo('/')
  } catch (err) {
    const status = (err as Response).status
    let message = 'Something went wrong.'

    if ([401, 404].includes(status)) {
      message = 'Invalid OTP. Please try again.'
    }

    toast.add({
      id: ToastId.VERIFY_SIGNUP_FAILED,
      color: 'red',
      title: 'Failed to verify email',
      description: message,
      icon: 'i-heroicons-exclamation-circle',
    })
  } finally {
    authStore.clearSignUpEmail()
    loading.value = false
  }
}
</script>
