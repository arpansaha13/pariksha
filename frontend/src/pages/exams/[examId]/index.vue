<template>
  <UCard
    v-if="exam"
    :ui="{
      header: 'flex items-center gap-2 justify-between',
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
        <EditableArea
          class="rounded-sm px-1 focus-within:outline hover:outline"
        >
          <EditablePreview as="h1" class="heading" />
          <EditableInput class="heading outline-none" />
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
      <h3 class="font-medium">Start date:</h3>
      <UPopover>
        <UButton color="neutral" variant="subtle" icon="i-lucide-calendar">
          <ClientOnly>
            {{ df.format(startsAt.toDate(getLocalTimeZone())) }}

            <template #fallback>
              <USkeleton class="h-4 w-[11ch] bg-neutral-200" />
            </template>
          </ClientOnly>
        </UButton>

        <template #content>
          <UCalendar v-model="startsAt" :min-value="today" class="p-2" />
        </template>
      </UPopover>
    </div>

    <div class="flex items-center gap-2">
      <h3 class="font-medium">End date:</h3>
      <UPopover>
        <UButton color="neutral" variant="subtle" icon="i-lucide-calendar">
          <ClientOnly>
            {{ df.format(endsAt.toDate(getLocalTimeZone())) }}

            <template #fallback>
              <USkeleton class="h-4 w-[11ch] bg-neutral-200" />
            </template>
          </ClientOnly>
        </UButton>

        <template #content>
          <UCalendar v-model="endsAt" :min-value="startsAt" class="p-2" />
        </template>
      </UPopover>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import {
  CalendarDateTime,
  DateFormatter,
  getLocalTimeZone,
} from '@internationalized/date'
import { debounceFilter } from '@vueuse/core'

const route = useRoute()
const examId = parseInt(route.params.examId as string)
const { data: exam } = await useExam(examId)

const fullCurrentUrl = ref('')
const { copy, copied, isSupported } = useClipboard({ source: fullCurrentUrl })

if (isSupported.value) {
  const url = useRequestURL()
  fullCurrentUrl.value = url.toString()
}

const df = new DateFormatter('en-US', { dateStyle: 'medium' })

const editableExamTitle = ref(exam.value!.title)
const startsAt = shallowRef(toCalendarDateTime(exam.value!.starts_at))
const endsAt = shallowRef(toCalendarDateTime(exam.value!.ends_at))

const date = new Date()
const today = new CalendarDateTime(
  date.getFullYear(),
  date.getMonth() + 1,
  date.getDate()
)

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

function addEndOfDayTime(date: CalendarDateTime) {
  return date.add({ hours: 23, minutes: 59, seconds: 59 })
}

// Sync with exam data
watch(exam, newExam => {
  if (!newExam) return
  editableExamTitle.value = newExam.title
  startsAt.value = toCalendarDateTime(newExam.starts_at)
  endsAt.value = toCalendarDateTime(newExam.ends_at)
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
</script>
