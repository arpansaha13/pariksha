<template>
  <main>
    <UCard>
      <template #header>
        <h1 class="heading">Create exam</h1>
      </template>

      <div v-if="hasNoPaper" class="flex flex-col items-center">
        <div class="flex items-center">
          <Icon name="i-heroicons-document-plus" size="2.5rem" />
        </div>

        <p class="mt-2 max-w-sm text-center text-pretty text-gray-500">
          You need to create or have access to a paper for creating an exam.
        </p>

        <UButton
          to="/papers/new"
          label="Create a paper"
          color="primary"
          variant="solid"
          class="mt-4"
        />
      </div>

      <UForm
        v-else
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
            class="w-52"
          >
            <template #item-label="{ item }">
              <div class="px-1">
                <p class="font-medium">{{ item.title }}</p>

                <p class="space-x-1 text-gray-500">
                  <span class="inline-block">
                    {{ getTotalQuestionsCountText(item.question_counts) }}
                  </span>

                  <template v-if="item.duration_minutes > 0">
                    <Dot />
                    <span class="inline-block">
                      {{ item.duration_minutes }} minutes
                    </span>
                  </template>
                </p>
              </div>
            </template>
          </USelectMenu>
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

      <template v-if="!hasNoPaper" #footer>
        <UButton
          color="primary"
          variant="solid"
          :loading="isLoading"
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
import { isNullOrUndefined } from '@arpansaha13/utils'

const newExamStore = useNewExamStore()

interface ExamFormState extends Pick<Exam, 'title' | 'type'> {
  paper_id: PaperId | undefined
  duration_hours: number
  duration_minutes: number
}

const formState = reactive<ExamFormState>({
  title: newExamStore.title ?? '',
  type: newExamStore.type ?? ExamAccessType.LINK,
  paper_id: newExamStore.paper_id ?? undefined,
  duration_hours: newExamStore.duration_hours ?? 0,
  duration_minutes: newExamStore.duration_minutes ?? 0,
})

const { data: papers } = await usePapers()
const hasNoPaper = ref(
  isNullOrUndefined(papers.value) || papers.value.length === 0
)

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

watch(
  () => formState.paper_id,
  paperId => {
    if (!paperId || !papers.value) return

    const selectedPaper = papers.value.find(p => p.id === paperId)
    if (!selectedPaper?.duration_minutes) return

    formState.duration_hours = Math.floor(selectedPaper.duration_minutes / 60)
    formState.duration_minutes = selectedPaper.duration_minutes % 60
  }
)

const df = new DateFormatter('en-US', {
  dateStyle: 'medium',
})

const isLoading = ref(false)

onBeforeUnmount(() => {
  // If not submitted, store the form values
  if (!isLoading.value) {
    newExamStore.title = formState.title
    newExamStore.type = formState.type
    newExamStore.paper_id = formState.paper_id ?? null
    newExamStore.startDate = startDate.value
    newExamStore.endDate = endDate.value
    newExamStore.duration_hours = formState.duration_hours
    newExamStore.duration_minutes = formState.duration_minutes
  }
})

const submitButtonRef = useTemplateRef('submitButton')
async function onSubmit() {
  // Do not end loading to keep button disabled till navigation
  isLoading.value = true

  // Consider entire day on end date
  const endsAt = endDate.value.add({ hours: 23, minutes: 59, seconds: 59 })
  const createdExam = await createExam({
    title: formState.title,
    type: formState.type,
    paper_id: formState.paper_id!,
    starts_at: startDate.value.toDate(getLocalTimeZone()),
    ends_at: endsAt.toDate(getLocalTimeZone()),
    duration_minutes: convertToMinutes(
      formState.duration_hours,
      formState.duration_minutes
    ),
  })

  newExamStore.$reset()
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

  const durationMinutes = convertToMinutes(
    formState.duration_hours,
    formState.duration_minutes
  )
  if (durationMinutes === 0) {
    errors.push({
      name: 'duration',
      message: 'Please set a duration for the exam',
    })
  } else if (durationMinutes > MAX_EXAM_DURATION_MINUTES) {
    errors.push({
      name: 'duration',
      message: 'Exam duration cannot be more than 24 hours',
    })
  }

  return errors
}

function getTotalQuestionsCountText(counts: PaperQuestionCounts) {
  const count = countTotalQuestions(counts)

  return `${count} ${count === 1 ? 'question' : 'questions'}`
}
</script>
