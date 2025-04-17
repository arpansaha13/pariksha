<template>
  <div>Exam {{ examId }} attempt</div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['check-exam-participant'],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)
</script>
