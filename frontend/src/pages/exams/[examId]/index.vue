<template>
  <ExamViewOwner v-if="isOwner" />

  <UContainer v-else class="max-w-3xl py-6">
    <ExamViewParticipant />
  </UContainer>
</template>

<script setup lang="ts">
import { ExamPermission } from '~/types/exam'

definePageMeta({
  middleware: ['check-exam-access'],
})

const nuxtApp = useNuxtApp()
const examAccess = nuxtApp.$examAccess as Awaited<
  ReturnType<typeof checkExamAccess>
>
const isOwner = examAccess.access_type === ExamPermission.OWNER
</script>
