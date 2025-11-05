<template>
  <div class="flex h-screen w-screen flex-col items-center justify-center">
    <UError
      :error="{
        statusCode: HttpStatus.NOT_FOUND,
        statusMessage: 'Page not found',
        message,
      }"
      redirect="/"
      :clear="{
        variant: 'subtle',
        color: 'primary',
        label: 'Go back to Home',
      }"
    />
  </div>
</template>

<script setup lang="ts">
defineProps({
  message: {
    type: String,
    default: 'Sorry, the page you are looking for does not exist.',
  },
})
</script>
