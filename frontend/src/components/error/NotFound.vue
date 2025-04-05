<template>
  <div class="flex h-screen w-screen flex-col items-center justify-center">
    <p class="text-primary-600 dark:text-primary-400 text-center font-bold">
      {{ HttpStatus.NOT_FOUND }}
    </p>
    <h1
      class="mt-2 text-center text-4xl font-bold text-gray-800 dark:text-white"
    >
      Page not found
    </h1>
    <p class="mt-2 text-center text-gray-600 dark:text-gray-400">
      {{ message }}
    </p>
    <div class="mt-6 flex justify-center">
      <UButton variant="subtle" color="primary" @click="redirectToHome">
        Go back to Home
      </UButton>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps({
  message: {
    type: String,
    default: 'Sorry, the page you are looking for does not exist.',
  },
})

function redirectToHome() {
  clearError({ redirect: '/' })
}
</script>
