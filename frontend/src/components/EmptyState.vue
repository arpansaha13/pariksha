<template>
  <div class="flex flex-col items-center">
    <div class="mt-1 text-gray-600">
      <Icon :name="icon" size="2.5rem" />
    </div>
    <p class="text-gray-500">{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon: string
  description: string
}>()
</script>
