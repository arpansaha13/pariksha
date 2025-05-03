<template>
  <UContainer v-if="isParticipant" class="max-w-3xl py-6">
    <ExamViewParticipant :exam-permission="examPermission!" />
  </UContainer>

  <ExamViewOwner v-else />
</template>

<script setup lang="ts">
import type { ExamPermission } from '~/types/exam'

definePageMeta({
  middleware: [
    'check-exam-permission',
    to => {
      const examId = parseInt(to.params.examId as string)
      const { data: examPermission } = useNuxtData<ExamPermission>(
        AsyncDataKeys.EXAM_ACCESS(examId)
      )
      if (examPermission.value!.can_participate) {
        setPageLayout('blank')
      }
    },
  ],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)

const { data: examPermission } = useNuxtData<ExamPermission>(
  AsyncDataKeys.EXAM_ACCESS(examId)
)

const isParticipant = examPermission.value!.can_participate
</script>
