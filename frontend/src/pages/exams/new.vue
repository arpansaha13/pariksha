<template>
  <main>
    <h1 class="mb-6 text-3xl font-bold">Create exam</h1>

    <UCard>
      <UForm
        :state="formState"
        class="flex flex-col gap-y-5"
        @submit.prevent="onSubmit"
      >
        <UFormField label="Type" required>
          <UInput v-model="formState.title" required />
        </UFormField>

        <UFormField
          label="Paper"
          description="Choose the paper from which the questions will be taken."
          required
        >
          <USelectMenu
            v-if="papers"
            v-model="selectedPaperId"
            :items="papers"
            value-key="id"
            label-key="title"
            required
            class="w-48"
          />
        </UFormField>

        <UFormField
          label="Start date"
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
import {
  CalendarDateTime,
  DateFormatter,
  getLocalTimeZone,
} from '@internationalized/date'
import { type Exam, ExamAccessType } from '~/types/exam'

const formState = reactive<Pick<Exam, 'title' | 'type'>>({
  title: '',
  type: ExamAccessType.LINK,
})

const { data: papers } = await usePapers()
const selectedPaperId = ref()

const date = new Date()
const today = new CalendarDateTime(
  date.getFullYear(),
  date.getMonth() + 1, // getMonth() is 0-indexed
  date.getDate()
)

const startDate = shallowRef(today)
const endDate = shallowRef(today)

const df = new DateFormatter('en-US', {
  dateStyle: 'medium',
})

const submitButtonRef = useTemplateRef('submitButton')
async function onSubmit() {
  // Consider entire day on end date
  const endsAt = endDate.value.add({ hours: 23, minutes: 59, seconds: 59 })

  const createdExam = await createExam({
    title: formState.title,
    type: formState.type,
    paper_id: selectedPaperId.value,
    starts_at: startDate.value.toDate(getLocalTimeZone()),
    ends_at: endsAt.toDate(getLocalTimeZone()),
  })

  await navigateTo(`/exams/${createdExam.id}`)
}
</script>
