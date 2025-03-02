<template>
  <div class="flex w-full max-w-md flex-col items-center">
    <div>
      <Logo class="mx-auto size-20" />
      <h1 class="mb-6 mt-4 text-center text-2xl font-bold">
        Login to your account
      </h1>
    </div>

    <UCard class="mt-6 w-full">
      <UForm :state="loginFormData" @submit="onSubmit">
        <div class="space-y-4">
          <UFormGroup label="Email" name="email">
            <UInput
              v-model="loginFormData.email"
              type="email"
              placeholder="Enter your email"
              autocomplete="email"
            />
          </UFormGroup>

          <UFormGroup label="Password" name="password">
            <UInput
              v-model="loginFormData.password"
              type="password"
              placeholder="Enter your password"
              autocomplete="current-password"
            />
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

const loginFormData = useState(() => ({
  email: '',
  password: '',
}))

const loading = useState(() => false)

async function onSubmit(e: SubmitEvent) {
  e.preventDefault()

  try {
    loading.value = true
    await login(loginFormData.value)

    // if (error.value) {
    //   throw error.value
    // }
    // console.log(data)

    await navigateTo('/')
  } catch (err) {
    console.error('Login failed:', err)
  } finally {
    loading.value = false
  }
}
</script>
