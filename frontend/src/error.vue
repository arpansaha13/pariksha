<template>
  <NuxtLoadingIndicator
    color="repeating-linear-gradient(to right,#00dc82 0%,#34cdfe 50%,#0047e1 100%)"
  />

  <ErrorNotFound
    v-if="error.statusCode === HttpStatus.NOT_FOUND"
    :message="error.message"
  />
  <ErrorForbidden
    v-else-if="error.statusCode === HttpStatus.FORBIDDEN"
    :message="error.message"
  />
</template>

<script setup lang="ts">
import type { NuxtError } from '#app'

defineProps({
  error: {
    type: Object as PropType<NuxtError>,
    default: null,
  },
})
</script>
