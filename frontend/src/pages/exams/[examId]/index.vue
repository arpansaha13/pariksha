<template>
  <div>Exam {{ examId }}</div>
</template>

<script setup lang="ts">
const route = useRoute()
const examId = parseInt(route.params.examId as string)
</script>
