<template>
  <div class="flex w-full max-w-sm flex-col items-center gap-6">
    <div>
      <Logo class="mx-auto size-20" />

      <h1 class="mb-2 mt-4 text-center text-2xl font-bold">Welcome back!</h1>

      <p class="text-center">
        <span>Don't have an account?</span>
        <ULink
          to="/auth/signup"
          class="text-primary-500 hover:text-primary-600 font-semibold"
        >
          Sign up<span class="font-normal text-black">.</span>
        </ULink>
      </p>
    </div>

    <UCard class="w-full">
      <UForm :state="loginFormData" @submit.prevent="onSubmit">
        <div class="space-y-4">
          <UFormGroup label="Email" name="email" required>
            <UInput
              v-model="loginFormData.email"
              type="email"
              placeholder="Enter your email"
              autocomplete="email"
              required
            />
          </UFormGroup>

          <UFormGroup label="Password" name="password" required>
            <template #hint>
              <ULink
                to="/auth/forgot-password"
                class="text-primary-500 hover:text-primary-600 font-medium"
              >
                Forgot password?
              </ULink>
            </template>
            <template #default>
              <UInput
                v-model="loginFormData.password"
                type="password"
                placeholder="Enter your password"
                autocomplete="current-password"
                required
              />
            </template>
          </UFormGroup>

          <UButton type="submit" color="primary" block :loading="loading">
            Login
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
  title: 'Login',
})

const toast = useToast()

const loginFormData = useState(() => ({
  email: '',
  password: '',
}))

const loading = useState(() => false)

async function onSubmit() {
  try {
    loading.value = true
    await login(loginFormData.value)
    await navigateTo('/')
  } catch {
    toast.add({
      id: ToastId.LOGIN_FAILED,
      color: 'red',
      title: 'Failed to login',
      description: 'Invalid email or password.',
      icon: 'i-heroicons-exclamation-circle',
    })
  } finally {
    loading.value = false
  }
}
</script>
