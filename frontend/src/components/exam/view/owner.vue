<template>
  <main class="space-y-6">
    <UCard
      v-if="exam"
      :ui="{
        header: 'flex items-center justify-between',
        body: 'space-y-4',
      }"
    >
      <template #header>
        <EditableRoot
          v-model="editableExamTitle"
          name="exam-title"
          activation-mode="focus"
          submit-mode="both"
          placeholder=""
          @submit="updateExamTitle"
        >
          <EditableArea class="rounded-sm focus-within:outline hover:outline">
            <EditablePreview as="h1" class="heading" />
            <EditableInput class="text-2xl font-bold outline-none" />
          </EditableArea>
        </EditableRoot>

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
          {{ isExamStarted ? 'Started at:' : 'Starts at:' }}
        </p>

        <DisplayDate
          v-if="isExamStarted"
          :date="startsAt"
          :df="df"
          :ui="{ skeleton: 'h-4 w-[11ch] bg-neutral-200' }"
        />

        <UPopover v-else>
          <UButton color="neutral" variant="subtle" icon="i-lucide-calendar">
            <DisplayDate
              :date="startsAt"
              :df="df"
              :ui="{ skeleton: 'h-4 w-[11ch] bg-neutral-200' }"
            />
          </UButton>

          <template #content>
            <UCalendar v-model="startsAt" :min-value="now" class="p-2" />
          </template>
        </UPopover>
      </div>

      <div class="flex items-center gap-2">
        <p class="font-medium">
          {{ isExamEnded ? 'Ended at:' : 'Ends at:' }}
        </p>

        <DisplayDate
          v-if="isExamEnded"
          :date="endsAt"
          :df="df"
          :ui="{ skeleton: 'h-4 w-[11ch] bg-neutral-200' }"
        />

        <UPopover v-else>
          <UButton color="neutral" variant="subtle" icon="i-lucide-calendar">
            <DisplayDate
              :date="endsAt"
              :df="df"
              :ui="{ skeleton: 'h-4 w-[11ch] bg-neutral-200' }"
            />
          </UButton>

          <template #content>
            <UCalendar v-model="endsAt" :min-value="startsAt" class="p-2" />
          </template>
        </UPopover>
      </div>

      <div class="flex items-center gap-2">
        <p class="font-medium">Duration:</p>
        <p>{{ formatDurationMinutes(exam.duration_minutes) }}</p>

        <UModal
          v-if="!isExamStarted"
          title="Edit duration"
          description="Edit the exam duration"
          @after:leave="updateExamDuration"
        >
          <UTooltip text="Edit duration">
            <UButton
              icon="lucide:edit"
              size="sm"
              color="neutral"
              variant="subtle"
            />
          </UTooltip>

          <template #body>
            <div class="grid grid-cols-2 gap-x-4">
              <UFormField label="Hours" name="duration_hours">
                <UInputNumber v-model="hours" :min="0" :max="24" />
              </UFormField>

              <UFormField label="Minutes" name="duration_minutes">
                <UInputNumber v-model="minutes" :min="0" :max="59" />
              </UFormField>
            </div>
          </template>
        </UModal>
      </div>
    </UCard>

    <UCard
      :ui="{
        header: 'flex items-center justify-between',
        body: '!py-2',
      }"
    >
      <template #header>
        <h2 class="heading">Participants</h2>
      </template>

      <UTable
        :data="participantsData ?? undefined"
        :loading="participantsPending"
        :columns="columns"
        class="flex-1"
      />
    </UCard>
  </main>
</template>

<script setup lang="ts">
import {
  type CalendarDateTime,
  DateFormatter,
  getLocalTimeZone,
} from '@internationalized/date'
import type { TableColumn } from '@nuxt/ui'
import { debounceFilter } from '@vueuse/core'
import { isNullOrUndefined } from '@arpansaha13/utils'
import {
  ExamParticipantStatus,
  type ExamParticipantResponse,
  type ExamPermission,
} from '~/types'

const UButton = resolveComponent('UButton')
const UBadge = resolveComponent('UBadge')

const route = useRoute()
const examId = parseInt(route.params.examId as string)

const { data: examPermission } = useNuxtData<ExamPermission>(
  AsyncDataKeys.EXAM_PERMISSION(examId)
)

const [
  { data: exam },
  { data: participantsData, pending: participantsPending },
] = await Promise.all([useExam(examId), useExamParticipants(examId)])

const fullCurrentUrl = ref('')
const { copy, copied, isSupported } = useClipboard({ source: fullCurrentUrl })

if (isSupported.value) {
  const url = useRequestURL()
  fullCurrentUrl.value = url.toString()
}

const now = toCalendarDateTime(new Date())
const editableExamTitle = ref(exam.value!.title)
const startsAt = shallowRef(toCalendarDateTime(exam.value!.starts_at))
const endsAt = shallowRef(toCalendarDateTime(exam.value!.ends_at))

const isExamStarted = isCalendarBefore(startsAt, now)
const isExamEnded = isCalendarBefore(endsAt, now)

const df = new DateFormatter('en-US', {
  dateStyle: 'long',
  timeStyle: 'short',
})

async function updateExamTitle() {
  editableExamTitle.value = editableExamTitle.value.trim()
  if (!editableExamTitle.value) {
    editableExamTitle.value = 'Untitled Exam'
  }
  if (editableExamTitle.value !== exam.value!.title) {
    try {
      await updateExam(examId, { title: editableExamTitle.value })
    } catch {
      // Rollback to original title on error
      editableExamTitle.value = exam.value!.title
    }
  }
}

// ______________________UPDATE START/END TIME______________________

function addEndOfDayTime(date: CalendarDateTime) {
  return date.add({ hours: 23, minutes: 59, seconds: 59 })
}

// Sync with exam data
watch(exam, newExam => {
  if (isNullOrUndefined(newExam)) return
  editableExamTitle.value = newExam.title
  ignoreStartsAtUpdates(() => {
    startsAt.value = toCalendarDateTime(newExam.starts_at)
  })
  ignoreEndsAtUpdates(() => {
    endsAt.value = toCalendarDateTime(newExam.ends_at)
  })
})

const WATCH_DEBOUNCE_MS = 700

const { ignoreUpdates: ignoreStartsAtUpdates } = watchIgnorable(
  startsAt,
  async newStartsAt => {
    try {
      const body: Parameters<typeof updateExam>[1] = {
        starts_at: startsAt.value.toDate(getLocalTimeZone()),
      }

      // If start date exceeds end date, update end date as well
      if (newStartsAt > endsAt.value) {
        ignoreEndsAtUpdates(() => {
          endsAt.value = newStartsAt
        })
        endsAt.value = newStartsAt
        body.ends_at = addEndOfDayTime(newStartsAt).toDate(getLocalTimeZone())
      }

      await updateExam(examId, body)
    } catch {
      // Rollback on error
      ignoreStartsAtUpdates(() => {
        startsAt.value = toCalendarDateTime(exam.value!.starts_at)
      })
      ignoreEndsAtUpdates(() => {
        endsAt.value = toCalendarDateTime(exam.value!.ends_at)
      })
    }
  },
  {
    eventFilter: debounceFilter(WATCH_DEBOUNCE_MS),
  }
)

const { ignoreUpdates: ignoreEndsAtUpdates } = watchIgnorable(
  endsAt,
  async newEndsAt => {
    try {
      const endsAtDate = addEndOfDayTime(newEndsAt).toDate(getLocalTimeZone())
      await updateExam(examId, { ends_at: endsAtDate })
    } catch {
      // Rollback on error
      ignoreEndsAtUpdates(() => {
        endsAt.value = toCalendarDateTime(exam.value!.ends_at)
      })
    }
  },
  { eventFilter: debounceFilter(WATCH_DEBOUNCE_MS) }
)

// _________________________DURATION UPDATE_________________________

const hours = ref(0)
const minutes = ref(0)

watchEffect(() => {
  if (exam.value && exam.value.duration_minutes) {
    hours.value = Math.floor(exam.value.duration_minutes / 60)
    minutes.value = exam.value.duration_minutes % 60
  }
})

function updateExamDuration() {
  const totalMinutes = hours.value * 60 + minutes.value
  if (totalMinutes > 0) {
    updateExam(exam.value!.id, { duration_minutes: totalMinutes })
  }
}

// ___________________EXAM PARTICIPANT TABLE DATA___________________

const participantStatusColors = {
  [ExamParticipantStatus.UNATTENDED]: 'error',
  [ExamParticipantStatus.INVITED]: 'info',
  [ExamParticipantStatus.STARTED]: 'success',
  [ExamParticipantStatus.ENDED]: 'warning',
  [ExamParticipantStatus.EVALUATED]: 'success',
} as const

const participantStatusText = {
  [ExamParticipantStatus.UNATTENDED]: 'Unattended',
  [ExamParticipantStatus.INVITED]: 'Invited',
  [ExamParticipantStatus.STARTED]: 'Started',
  [ExamParticipantStatus.ENDED]: 'Pending evaluation',
  [ExamParticipantStatus.EVALUATED]: 'Evaluated',
} as const

const columns: TableColumn<ExamParticipantResponse>[] = [
  {
    header: 'Name',
    cell: ({ row }) => {
      const participantId = row.original.id
      const name = row.original.first_name + ' ' + row.original.last_name

      if (
        !examPermission.value?.can_evaluate ||
        [
          ExamParticipantStatus.UNATTENDED,
          ExamParticipantStatus.INVITED,
          ExamParticipantStatus.STARTED,
        ].includes(row.original.status)
      ) {
        return name
      }

      return h(UButton, {
        label: name,
        to: `/exams/${examId}/evaluation/${participantId}`,
        variant: 'link',
        ui: { base: 'px-0' },
      })
    },
  },
  {
    accessorKey: 'email',
    header: 'Email',
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => {
      const statusValue = row.getValue<ExamParticipantStatus>('status')
      const color = participantStatusColors[statusValue]

      return h(
        UBadge,
        { class: 'capitalize', variant: 'subtle', color },
        () => participantStatusText[statusValue]
      )
    },
  },
  {
    header: 'Score',
    cell: ({ row }) => {
      if (row.original.status !== ExamParticipantStatus.EVALUATED) return '--'

      return h(
        UBadge,
        {
          variant: 'subtle',
          color: getScoreColor(
            row.original.score_awarded,
            exam.value!.max_score
          ),
        },
        () => `${row.original.score_awarded} / ${exam.value!.max_score}`
      )
    },
  },
]
</script>
