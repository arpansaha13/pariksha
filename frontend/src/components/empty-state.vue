<template>
  <div class="flex flex-col items-center text-gray-500">
    <div class="mt-1 text-gray-600">
      <Icon :name="icon" size="2.5rem" />
    </div>
    <p v-if="title" class="text-sm font-medium">{{ title }}</p>
    <p v-if="description" class="text-center text-sm text-pretty">
      {{ description }}
    </p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon: string
  title?: string
  description?: string
}>()
</script>
