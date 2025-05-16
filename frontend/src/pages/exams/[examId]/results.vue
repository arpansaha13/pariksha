<template>
  <main class="space-y-4">
    <UCard>
      <template #header>
        <h2 class="heading">{{ examData?.title }} - Results</h2>
      </template>

      <div>
        <p v-if="isParticipantExamEvaluated">
          Total score: {{ totalScoreObtained }} / {{ totalScore }}
        </p>
        <p v-else>
          <EmptyState
            icon="lucide:file-clock"
            description="Your exam is under evaluation — results coming soon!"
          />
        </p>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h2 class="heading">My answers</h2>
      </template>

      <UAccordion
        type="multiple"
        :items="resultsData ?? undefined"
        label-key="question.content.statement"
      >
        <template #body="{ item: result }">
          <div v-if="isNullOrUndefined(result.answer.content)">
            <EmptyState
              icon="lucide:file-question"
              description="This question is unanswered"
            />
          </div>
          <URadioGroup
            v-else-if="result.type === QuestionType.MCQ"
            :default-value="result.answer.content?.optionIndex"
            :items="mcqQuestionOptionsMap[result.question.id]"
            variant="card"
            disabled
            :ui="{
              wrapper: 'ml-3',
              fieldset: 'space-y-1',
              label: 'opacity-100',
              item: 'opacity-100',
            }"
          />
          <p v-else>
            {{ result.answer.content?.text }}
          </p>
        </template>

        <template #trailing="{ item, open }">
          <div class="ml-auto flex items-center gap-3">
            <p v-if="isParticipantExamEvaluated">
              {{ item.answer.score_awarded }} / {{ item.question.max_score }}
            </p>
            <Icon
              name="i-lucide:chevron-down"
              size="1.25rem"
              :class="[
                'transition-transform duration-300',
                open && 'rotate-180',
              ]"
            />
          </div>
        </template>
      </UAccordion>
    </UCard>
  </main>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'
import type { RadioGroupItem } from '@nuxt/ui'
import {
  type ExamPermission,
  ExamParticipantStatus,
  QuestionType,
} from '~/types'

definePageMeta({
  middleware: [
    'check-exam-permission',
    to => {
      const examId = parseInt(to.params.examId as string)
      const { data: examPermission } = useNuxtData<ExamPermission>(
        AsyncDataKeys.EXAM_PERMISSION(examId)
      )
      if (!examPermission.value!.can_participate) {
        throw abortNavigation({
          statusCode: HttpStatus.FORBIDDEN,
          message: 'You do not have access to this page.',
        })
      }
    },
  ],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)

const { data: examPermission } = useNuxtData<ExamPermission>(
  AsyncDataKeys.EXAM_PERMISSION(examId)
)
const isParticipantExamEvaluated =
  examPermission.value!.participant_status === ExamParticipantStatus.EVALUATED

const [{ data: examData }, { data: resultsData }] = await Promise.all([
  useExam(examId),
  useExamResults(examId),
])

const totalScore = computed(() => {
  if (isNullOrUndefined(resultsData.value)) return 0

  return resultsData.value.reduce(
    (acc, result) => acc + result.question.max_score,
    0
  )
})

const totalScoreObtained = computed(() => {
  if (isNullOrUndefined(resultsData.value)) return 0

  return resultsData.value.reduce(
    (acc, result) => acc + result.answer.score_awarded,
    0
  )
})

const mcqQuestionOptionsMap = shallowRef<Record<number, RadioGroupItem[]>>({})

if (!isNullOrUndefined(resultsData.value)) {
  for (const result of resultsData.value) {
    if (result.type !== QuestionType.MCQ) continue

    mcqQuestionOptionsMap.value[result.question.id] =
      result.question.content.options.map((option, i) => ({
        value: i,
        label: option,
      }))
  }
}
</script>
