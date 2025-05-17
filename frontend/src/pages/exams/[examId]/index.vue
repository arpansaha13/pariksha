<template>
  <ExamViewParticipant
    v-if="isParticipant"
    :exam-permission="examPermission!"
  />

  <ExamViewOwner v-else />
</template>

<script setup lang="ts">
import type { ExamPermission } from '~/types/exam'

definePageMeta({
  middleware: ['check-exam-permission'],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)

const { data: examPermission } = useNuxtData<ExamPermission>(
  AsyncDataKeys.EXAM_PERMISSION(examId)
)

const isParticipant = examPermission.value!.can_participate
</script>
