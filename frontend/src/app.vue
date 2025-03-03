<template>
  <NuxtLayout>
    <NuxtLoadingIndicator
      color="repeating-linear-gradient(to right,#00dc82 0%,#34cdfe 50%,#0047e1 100%)"
    />
    <NuxtPage />
    <UNotifications
      :ui="{ position: 'top-0 bottom-[unset] left-1/2 -translate-x-1/2' }"
    />
  </NuxtLayout>
</template>

<script setup lang="ts">
useHead({
  titleTemplate: titleChunk => {
    return titleChunk ? `${titleChunk} - Pariksha` : 'Pariksha'
  },
})
</script>
