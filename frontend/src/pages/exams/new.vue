<template>
  <main>
    <h1 class="mb-6 text-3xl font-bold">Create exam</h1>

    <UCard>
      <UForm
        :state="formState"
        :validate="validate"
        :validate-on="[]"
        class="flex flex-col gap-y-5"
        @submit.prevent="onSubmit"
      >
        <UFormField label="Title" name="type" required>
          <UInput v-model="formState.title" required />
        </UFormField>

        <UFormField
          label="Paper"
          name="paper_id"
          description="Choose the paper from which the questions will be taken."
          required
        >
          <USelectMenu
            v-if="papers"
            v-model="formState.paper_id"
            :items="papers"
            value-key="id"
            label-key="title"
            required
            class="w-48"
          />
        </UFormField>

        <UFormField
          label="Start date"
          name="start_date"
          description="Candidates will be able to take their exam from this date."
          required
        >
          <UPopover>
            <UButton color="neutral" variant="subtle" icon="i-lucide-calendar">
              <ClientOnly placeholder="Select a date">
                {{
                  startDate
                    ? df.format(startDate.toDate(getLocalTimeZone()))
                    : 'Select a date'
                }}
              </ClientOnly>
            </UButton>

            <template #content>
              <UCalendar v-model="startDate" :min-value="today" class="p-2" />
            </template>
          </UPopover>
        </UFormField>

        <UFormField
          label="End date"
          name="end_date"
          description="Candidates will not be able to take their exam after this date."
          required
        >
          <UPopover>
            <UButton color="neutral" variant="subtle" icon="i-lucide-calendar">
              <ClientOnly placeholder="Select a date">
                {{
                  endDate
                    ? df.format(endDate.toDate(getLocalTimeZone()))
                    : 'Select a date'
                }}
              </ClientOnly>
            </UButton>

            <template #content>
              <UCalendar v-model="endDate" :min-value="startDate" class="p-2" />
            </template>
          </UPopover>
        </UFormField>

        <UFormField
          label="Duration"
          description="How much time the candidate will have to complete the exam."
          name="duration"
          required
        >
          <div class="flex gap-x-4">
            <UFormField label="Hours" name="duration_hours">
              <UInputNumber
                v-model="formState.duration_hours"
                :min="0"
                :max="24"
              />
            </UFormField>

            <UFormField label="Minutes" name="duration_minutes">
              <UInputNumber
                v-model="formState.duration_minutes"
                :min="0"
                :max="59"
              />
            </UFormField>
          </div>
        </UFormField>

        <button ref="submitButton" type="submit" class="hidden" />
      </UForm>

      <template #footer>
        <UButton
          color="primary"
          variant="solid"
          loading-auto
          @click="submitButtonRef?.click()"
        >
          Create exam
        </UButton>
      </template>
    </UCard>
  </main>
</template>

<script setup lang="ts">
import type { FormError } from '@nuxt/ui'
import {
  type CalendarDateTime,
  DateFormatter,
  getLocalTimeZone,
} from '@internationalized/date'
import { type Exam, ExamAccessType } from '~/types/exam'
import { isNullOrUndefined } from '@arpansaha13/utils'

const newExamStore = useNewExamStore()

interface ExamFormState extends Pick<Exam, 'title' | 'type'> {
  paper_id: number | undefined
  duration_hours: number
  duration_minutes: number
}

const formState = reactive<ExamFormState>({
  title: newExamStore.title ?? '',
  type: newExamStore.type ?? ExamAccessType.LINK,
  paper_id: newExamStore.paper_id ?? undefined,
  duration_hours: 0,
  duration_minutes: 0,
})

const { data: papers } = await usePapers()

const today = toCalendarDateTime(new Date())

const startDate = shallowRef(
  (newExamStore.startDate as CalendarDateTime | null) ?? today
)
const endDate = shallowRef(
  (newExamStore.endDate as CalendarDateTime | null) ?? today
)

watch(startDate, newValue => {
  if (newValue > endDate.value) {
    endDate.value = newValue
  }
})

const df = new DateFormatter('en-US', {
  dateStyle: 'medium',
})

onBeforeUnmount(() => {
  newExamStore.title = formState.title
  newExamStore.type = formState.type
  newExamStore.paper_id = formState.paper_id ?? null
  newExamStore.startDate = startDate.value
  newExamStore.endDate = endDate.value
})

function calcMinutes(hours: number | undefined, minutes: number | undefined) {
  if (isNullOrUndefined(hours)) return minutes ?? 0
  if (isNullOrUndefined(minutes)) return hours * 60
  return hours * 60 + minutes
}

const submitButtonRef = useTemplateRef('submitButton')
async function onSubmit() {
  // Consider entire day on end date
  const endsAt = endDate.value.add({ hours: 23, minutes: 59, seconds: 59 })
  const createdExam = await createExam({
    title: formState.title,
    type: formState.type,
    paper_id: formState.paper_id!,
    starts_at: startDate.value.toDate(getLocalTimeZone()),
    ends_at: endsAt.toDate(getLocalTimeZone()),
    duration_minutes: calcMinutes(
      formState.duration_hours,
      formState.duration_minutes
    ),
  })

  newExamStore.clear()
  await navigateTo(`/exams/${createdExam.id}`)
}

function validate(formState: Partial<ExamFormState>): FormError[] {
  const errors: FormError[] = []

  if (!formState.paper_id) {
    errors.push({
      name: 'paper_id',
      message: 'Please select a paper',
    })
  }

  if (calcMinutes(formState.duration_hours, formState.duration_minutes) === 0) {
    errors.push({
      name: 'duration',
      message: 'Please set a duration for the exam',
    })
  }

  return errors
}
</script>
