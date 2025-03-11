<template>
  <UContainer
    as="main"
    :ui="{
      base: 'py-4 h-full grid flex-grow grid-cols-3 gap-6 grid-rows-[auto,_1fr]',
    }"
  >
    <div v-if="paper" class="col-span-2 flex items-center gap-2">
      <Icon name="i-heroicons-document-text" size="2rem" />
      <h1 class="text-xl font-semibold">{{ paper.title }}</h1>

      <UButton
        :to="{
          path: `/papers/${paperId}/edit`,
          query: route.query,
        }"
        label="Start editing"
        icon="i-heroicons-pencil-square"
        size="xs"
        color="white"
        variant="ghost"
        no-prefetch
        :ui="{ base: 'ml-auto' }"
      />

      <UButton
        to="/papers"
        label="Back"
        icon="i-heroicons-arrow-uturn-left"
        size="xs"
        color="white"
        variant="ghost"
      />
    </div>

    <div />

    <div class="col-span-2 flex h-full flex-col gap-y-4">
      <div
        v-if="categoryLinks !== null"
        class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800"
      >
        <UHorizontalNavigation :links="categoryLinks" />
      </div>

      <UCard v-if="question" :ui="{ base: 'flex-grow' }">
        <QuestionMcq
          v-if="question.type === QuestionType.MCQ"
          :question="question.question"
        />
        <QuestionNonMcq v-else :question="question.question" />
      </UCard>

      <UCard
        :ui="{ body: { base: 'flex', padding: 'px-4 py-4 sm:px-6 sm:py-5' } }"
      >
        <UButton
          v-if="prevQuestionId"
          color="white"
          label="Previous"
          :to="{ query: { ...route.query, question: prevQuestionId } }"
        />
        <UButton
          v-if="nextQuestionId"
          color="white"
          label="Next"
          :to="{ query: { ...route.query, question: nextQuestionId } }"
          :ui="{ base: 'ml-auto' }"
        />
      </UCard>
    </div>

    <div>
      <h2 class="mb-4 text-lg font-semibold">Question Pallet</h2>

      <UCard v-if="currentCategoryQuestions">
        <ul class="flex flex-wrap gap-4">
          <li v-for="(q, i) of currentCategoryQuestions" :key="q.id">
            <UButton
              :to="{ query: { ...route.query, question: q.id } }"
              :color="currentQuestionId === q.id ? 'primary' : 'white'"
              :variant="currentQuestionId === q.id ? 'outline' : 'solid'"
              size="lg"
              :ui="{
                base: 'size-10 flex items-center justify-center',
                rounded: 'rounded-full',
              }"
            >
              {{ i + 1 }}
            </UButton>
          </li>
        </ul>
      </UCard>
    </div>
  </UContainer>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionType } from '~/types'

definePageMeta({
  layout: 'cover',
})

const route = useRoute()
const paperId = parseInt(route.params.paperId as string)
const { data: paper } = await usePaper(paperId)
const { data: groupedQuestions } = await usePaperQuestions(paperId)
const { data: sortedCategories } = await usePaperCategories(paperId)

const lastVisitedQuestionForCategory = ref({})

watch(
  route,
  newRoute => {
    const query = newRoute.query
    if (isNullOrUndefined(query)) return
    lastVisitedQuestionForCategory.value[query.category] = query.question
  },
  { immediate: true }
)

function getQuestionIdForCategoryId(categoryId: number) {
  const categoryQuestions = groupedQuestions.value?.[categoryId]
  let questionId = lastVisitedQuestionForCategory.value[categoryId]
  if (!questionId) questionId = categoryQuestions?.[0].id.toString()
  return questionId
}

// Add initial `category` and `question` queries, if missing
if (!route.query.category && sortedCategories.value?.length) {
  const categoryId = sortedCategories.value[0].id.toString()
  const questionId = getQuestionIdForCategoryId(categoryId)
  const query = { category: categoryId, question: questionId }
  navigateTo({ query }, { replace: true })
}

const categoryLinks = computed(() => {
  if (!sortedCategories.value) return null

  return sortedCategories.value.map(category => ({
    label: category.name,
    to: {
      query: {
        category: category.id,
        question: getQuestionIdForCategoryId(category.id),
      },
    },
    exactQuery: true,
  }))
})

const currentCategoryId = computed(() => {
  return route.query.category ? parseInt(route.query.category as string) : null
})

const currentCategoryQuestions = computed(() => {
  if (!groupedQuestions.value || !currentCategoryId.value) return []
  return groupedQuestions.value[currentCategoryId.value] ?? []
})

const currentQuestionId = computed(() => {
  return route.query.question ? parseInt(route.query.question as string) : null
})

const currentQuestionIdx = computed(() => {
  if (!currentCategoryQuestions.value || !currentQuestionId.value) return -1
  return currentCategoryQuestions.value.findIndex(
    q => q.id === currentQuestionId.value
  )
})

const question = computed(() => {
  if (!currentQuestionIdx.value === -1) return null
  return currentCategoryQuestions.value?.[currentQuestionIdx.value] ?? null
})

const prevQuestionId = computed(() => {
  if (!currentCategoryQuestions.value || currentQuestionIdx.value <= 0)
    return null
  return currentCategoryQuestions.value[currentQuestionIdx.value - 1].id
})

const nextQuestionId = computed(() => {
  if (
    !currentCategoryQuestions.value ||
    currentQuestionIdx.value === -1 ||
    currentQuestionIdx.value >= currentCategoryQuestions.value.length - 1
  ) {
    return null
  }

  return currentCategoryQuestions.value[currentQuestionIdx.value + 1].id
})
</script>
