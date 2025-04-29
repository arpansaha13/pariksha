<template>
  <ExamViewOwner v-if="isOwner" />

  <UContainer v-else class="max-w-3xl py-6">
    <ExamViewParticipant :exam-access="examAccess!" />
  </UContainer>
</template>

<script setup lang="ts">
import { ExamPermission } from '~/types/exam'

definePageMeta({
  middleware: [
    'check-exam-access',
    to => {
      const examId = parseInt(to.params.examId as string)
      const { data: examAccess } = useNuxtData(
        AsyncDataKeys.EXAM_ACCESS(examId)
      )
      if (examAccess.value.access_type === ExamPermission.PARTICIPANT) {
        setPageLayout('blank')
      }
    },
  ],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)

const { data: examAccess } = useNuxtData(AsyncDataKeys.EXAM_ACCESS(examId))

const isOwner = examAccess.value!.access_type === ExamPermission.OWNER
</script>
