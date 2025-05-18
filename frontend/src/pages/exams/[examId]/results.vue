<template>
  <main class="space-y-4">
    <UCard>
      <template #header>
        <h1 class="heading">{{ examData?.title }} - Results</h1>
      </template>

      <div>
        <p
          v-if="
            isParticipantExamEvaluated &&
            !isNullOrUndefined(participant) &&
            !isNullOrUndefined(examData)
          "
          class="flex items-center gap-x-2"
        >
          <span class="inline-block font-medium">Total score:</span>
          <UBadge
            :color="
              getScoreColor(participant.score_awarded, examData.max_score)
            "
            variant="subtle"
          >
            {{ participant.score_awarded }} / {{ examData.max_score }}
          </UBadge>
        </p>
        <EmptyState
          v-else
          icon="lucide:file-clock"
          description="Your exam is under evaluation — results coming soon!"
        />
      </div>
    </UCard>

    <UCard :ui="{ body: '!pt-2' }">
      <template #header>
        <h2 class="heading">My answers</h2>
      </template>

      <ScrollAreaRoot
        v-if="!isNullOrUndefined(categoryLinks)"
        class="mb-2 flex items-center justify-between border-b border-gray-200 dark:border-gray-800"
      >
        <ScrollAreaViewport>
          <UNavigationMenu
            :items="categoryLinks"
            color="primary"
            orientation="horizontal"
            variant="link"
            highlight
          />
        </ScrollAreaViewport>
        <ScrollAreaScrollbar
          class="flex touch-none bg-white p-0.5 transition-colors ease-out select-none data-[orientation=horizontal]:h-2 data-[orientation=horizontal]:flex-col"
          orientation="horizontal"
        >
          <ScrollAreaThumb
            class="relative flex-1 rounded-sm bg-gray-200 transition-colors before:absolute before:top-1/2 before:left-1/2 before:h-full before:w-full before:-translate-x-1/2 before:-translate-y-1/2 before:content-[''] hover:bg-gray-300"
          />
        </ScrollAreaScrollbar>
      </ScrollAreaRoot>

      <UAccordion
        type="multiple"
        :items="currentCategoryQuestionAnswers ?? undefined"
        label-key="question.content.statement"
      >
        <template #body="{ item }">
          <div v-if="isNullOrUndefined(item.answer?.content)">
            <EmptyState
              icon="lucide:file-question"
              description="This question is unanswered"
            />
          </div>
          <URadioGroup
            v-else-if="item.type === QuestionType.MCQ"
            :default-value="item.answer.content?.optionIndex"
            :items="mcqQuestionOptionsMap![item.question.id]"
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
            {{ item.answer.content?.text }}
          </p>
        </template>

        <template #trailing="{ item, open }">
          <div class="ml-auto flex items-center gap-3">
            <UBadge
              v-if="isNullOrUndefined(item.answer)"
              color="error"
              variant="subtle"
            >
              Unanswered
            </UBadge>
            <UBadge
              v-else-if="
                isParticipantExamEvaluated && !isNullOrUndefined(resultsMap)
              "
              :color="
                getScoreColor(
                  resultsMap[item.answer.id].score_awarded,
                  item.question.max_score
                )
              "
              variant="subtle"
            >
              {{ resultsMap[item.answer.id].score_awarded }} /
              {{ item.question.max_score }}
            </UBadge>
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
  type Question,
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

const [
  { data: examData },
  { data: resultsMap },
  { data: participant },
  { data: sortedCategories },
] = await Promise.all([
  useExam(examId),
  useExamResults(examId),
  useExamParticipant(examId),
  useExamCategories(examId),
])

if (!route.query.category && sortedCategories.value?.length) {
  const catId = sortedCategories.value[0].id
  await navigateTo({ query: { category: catId } })
}

const { data: groupedQuestionAnswers } = await useExamParticipantAnswers(
  participant.value!.id
)

const currentCategoryQuestionAnswers = computed(() => {
  if (
    isNullOrUndefined(groupedQuestionAnswers.value) ||
    !route.query.category
  ) {
    return null
  }
  const catId = parseInt(route.query.category as string)
  return groupedQuestionAnswers.value[catId]
})

const categoryLinks = computed(() => {
  if (isNullOrUndefined(sortedCategories.value)) return null

  return sortedCategories.value.map(category => ({
    label: category.name,
    to: {
      query: { category: category.id },
    },
    exactQuery: true,
    replace: true,
  }))
})

const mcqQuestionOptionsMap = computed(() => {
  if (isNullOrUndefined(groupedQuestionAnswers.value)) return null

  const map: Record<Question['id'], RadioGroupItem[]> = {}
  for (const categoryQuestionAnswers of Object.values(
    groupedQuestionAnswers.value
  )) {
    for (const qa of categoryQuestionAnswers) {
      if (qa.type !== QuestionType.MCQ) continue
      map[qa.question.id] = qa.question.content.options.map((option, i) => ({
        value: i,
        label: option,
      }))
    }
  }

  return map
})
</script>
