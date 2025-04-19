<template>
  <UCard
    v-if="exam"
    :ui="{
      header: 'flex items-center gap-2 justify-between',
      body: 'space-y-4',
    }"
  >
    <template #header>
      <h1 class="heading">{{ exam.title }}</h1>

      <ClientOnly>
        <UTooltip v-if="isSupported" text="Copy link to exam">
          <UButton
            :icon="copied ? 'i-lucide-copy-check' : 'i-lucide-link'"
            size="sm"
            variant="outline"
            :color="copied ? 'primary' : 'neutral'"
            @click="() => copy()"
          />
        </UTooltip>
      </ClientOnly>
    </template>

    <div class="flex items-center gap-2">
      <p class="font-medium">
        {{ isCalendarBefore(startsAt, now) ? 'Started at:' : 'Starts at:' }}
      </p>

      <DisplayDate
        :date="startsAt"
        :df="df"
        :ui="{ skeleton: 'h-4 w-[26ch]' }"
      />
    </div>

    <div class="flex items-center gap-2">
      <p class="font-medium">
        {{ isCalendarBefore(endsAt, now) ? 'Ended at:' : 'Ends at:' }}
      </p>

      <DisplayDate :date="endsAt" :df="df" :ui="{ skeleton: 'h-4 w-[26ch]' }" />
    </div>

    <template
      v-if="isCalendarBefore(startsAt, now) && isCalendarAfter(endsAt, now)"
      #footer
    >
      <p v-if="isParticipantExamEnded">You have already attempted this exam</p>

      <UButton
        v-else-if="isParticipantExamStarted"
        label="Continue"
        :to="`/exams/${examId}/attempt`"
      />

      <UButton
        v-else
        label="Start exam"
        loading-auto
        @click="handleStartExam"
      />
    </template>
  </UCard>
</template>

<script setup lang="ts">
import { DateFormatter } from '@internationalized/date'
import { ExamParticipantStatus } from '~/types'

const nuxtApp = useNuxtApp()
const examAccess = nuxtApp.$examAccess as Awaited<
  ReturnType<typeof checkExamAccess>
>
const isParticipantExamStarted =
  examAccess.participant_status === ExamParticipantStatus.STARTED
const isParticipantExamEnded =
  examAccess.participant_status === ExamParticipantStatus.ENDED

const route = useRoute()
const examId = parseInt(route.params.examId as string)
const { data: exam } = await useExam(examId)

const fullCurrentUrl = ref('')
const { copy, copied, isSupported } = useClipboard({ source: fullCurrentUrl })

if (isSupported.value) {
  const url = useRequestURL()
  fullCurrentUrl.value = url.toString()
}

const df = new DateFormatter('en-US', {
  dateStyle: 'long',
  timeStyle: 'short',
})

const now = toCalendarDateTime(new Date())
const startsAt = toCalendarDateTime(exam.value!.starts_at)
const endsAt = toCalendarDateTime(exam.value!.ends_at)

async function handleStartExam() {
  await startExam(examId)
  await navigateTo(`/exams/${examId}/attempt`)
}
</script>
