<template>
  <button class="outline-hidden!" @click="toggleDark">
    <div class="dark:i-carbon:moon i-carbon:sun text-2xl" />
  </button>
</template>

<script setup lang="ts">
const color = useColorMode()
function toggleDark() {
  color.preference = color.value === 'dark' ? 'light' : 'dark'
}
</script>
