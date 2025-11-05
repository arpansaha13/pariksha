<template>
  <div class="flex h-screen w-screen flex-col items-center justify-center">
    <UError
      :error="{
        statusCode: HttpStatus.FORBIDDEN,
        statusMessage: 'Forbidden',
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
    default: 'You do not have access to this resource',
  },
})
</script>
