<template>
  <ExamViewOwner v-if="isOwner" />

  <UContainer v-else class="max-w-3xl py-6">
    <ExamViewParticipant :exam-access="examAccess!" />
  </UContainer>
</template>

<script setup lang="ts">
import { ExamPermission } from '~/types/exam'

definePageMeta({
  middleware: ['check-exam-access'],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)

const { data: examAccess } = await useExamCheckAccess(examId)

const isOwner = examAccess.value!.access_type === ExamPermission.OWNER
</script>
