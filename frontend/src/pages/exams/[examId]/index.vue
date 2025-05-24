<template>
  <ExamViewParticipant
    v-if="isParticipant"
    :exam-permission="examPermission!"
  />

  <ExamViewOwner v-else />
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['check-exam-permission'],
})

const route = useRoute()
const examId = route.params.examId as ExamId

const { data: examPermission } = useNuxtData<ExamPermission>(
  AsyncDataKeys.EXAM_PERMISSION(examId)
)

const isParticipant = examPermission.value!.can_participate
</script>
