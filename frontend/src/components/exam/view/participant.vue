<template>
  <main class="space-y-6">
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

        <DisplayDate
          :date="endsAt"
          :df="df"
          :ui="{ skeleton: 'h-4 w-[26ch]' }"
        />
      </div>
    </UCard>

    <ExamViewUpdateUserForm
      v-if="!isParticipantExamStarted && !isParticipantExamEnded"
      ref="updateUserForm"
      v-model:form-data="updateUserFormData"
      @submit="onStartExamSubmit"
    />

    <UCard>
      <template
        v-if="isCalendarBefore(startsAt, now) && isCalendarAfter(endsAt, now)"
      >
        <p v-if="isParticipantExamEnded">
          You have already attempted this exam
        </p>

        <UButton
          v-else-if="isParticipantExamStarted"
          label="Continue"
          :to="`/exams/${examId}/attempt`"
        />

        <UButton
          v-else
          label="Start exam"
          :loading="isStartingExam"
          @click="updateUserFormRef!.submit()"
        />
      </template>
    </UCard>
  </main>
</template>

<script setup lang="ts">
import { ExamViewUpdateUserForm } from '#components'
import { DateFormatter } from '@internationalized/date'
import type { ComponentExposed } from 'vue-component-type-helpers'
import { ExamParticipantStatus } from '~/types'

const props = defineProps({
  examAccess: {
    type: Object as PropType<
      NonNullable<ReturnType<typeof useExamCheckAccess>['data']['value']>
    >,
    required: true,
  },
})

const isParticipantExamStarted =
  props.examAccess.participant_status === ExamParticipantStatus.STARTED
const isParticipantExamEnded =
  props.examAccess.participant_status === ExamParticipantStatus.ENDED

const route = useRoute()
const examId = parseInt(route.params.examId as string)

const [{ data: exam }, { data: user }] = await Promise.all([
  useExam(examId),
  useAuthUser(),
])

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

// _________________________UPDATE USER_________________________
const updateUserFormRef =
  useTemplateRef<ComponentExposed<typeof ExamViewUpdateUserForm>>(
    'updateUserForm'
  )

const updateUserFormData = reactive({
  first_name: user.value!.first_name ?? '',
  last_name: user.value!.last_name ?? '',
})

function handleUpdateAuthUser() {
  const filteredUpdateUserFormData: Partial<typeof updateUserFormData> = {}

  if (updateUserFormData.first_name !== user.value!.first_name) {
    filteredUpdateUserFormData.first_name = updateUserFormData.first_name
  }

  if (updateUserFormData.last_name !== user.value!.last_name) {
    filteredUpdateUserFormData.last_name = updateUserFormData.last_name
  }

  if (Object.keys(filteredUpdateUserFormData).length > 0) {
    return updateAuthUser(filteredUpdateUserFormData)
  }
}

// _________________________START EXAM___________________________
const isStartingExam = ref(false)

async function onStartExamSubmit() {
  try {
    isStartingExam.value = true
    await handleUpdateAuthUser()
    await startExam(examId)
    await navigateTo(`/exams/${examId}/attempt`)
  } catch {
    // Stop loading in case of error
    // Otherwise let the loader run till page navigation
    isStartingExam.value = false
  }
}
</script>
