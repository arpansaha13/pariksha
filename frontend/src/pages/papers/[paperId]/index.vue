<template>
  <div>{{ paperId }}</div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'blank',
})

const route = useRoute()
const paperId = route.params.paperId as string
</script>
