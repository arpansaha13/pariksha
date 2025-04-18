<template>
  <div>Exam {{ examId }} attempt</div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'blank',
  middleware: ['check-exam-participant'],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)
const { data: exam } = await useExam(examId)
const { data: groupedQuestions } = await useExamQuestions(examId)
const { data: sortedCategories } = await useExamCategories(examId)

console.log('exam', exam)
console.log('groupedQuestions', groupedQuestions)
console.log('sortedCategories', sortedCategories)
</script>
