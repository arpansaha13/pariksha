<template>
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
      size="sm"
      color="neutral"
      variant="ghost"
      no-prefetch
      class="ml-auto"
    />

    <UButton
      to="/papers"
      label="Back"
      icon="i-heroicons-arrow-uturn-left"
      size="sm"
      color="neutral"
      variant="ghost"
    />
  </div>

  <div class="col-span-2 flex h-full flex-col gap-y-4">
    <ScrollAreaRoot
      v-if="categoryLinks !== null"
      class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800"
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
  </div>

  <div class="col-start-3 row-span-2 row-start-2">
    <h2 class="mb-4 text-lg font-semibold">Question Pallet</h2>

    <UCard v-if="currentCategoryQuestions">
      <ul class="flex flex-wrap gap-4">
        <li v-for="(q, i) of currentCategoryQuestions" :key="q.id">
          <UButton
            :to="{ query: { ...route.query, question: q.id } }"
            :color="currentQuestionId === q.id ? 'primary' : 'neutral'"
            :variant="currentQuestionId === q.id ? 'subtle' : 'outline'"
            size="lg"
            class="flex size-10 items-center justify-center rounded-full"
          >
            {{ i + 1 }}
          </UButton>
        </li>
      </ul>
    </UCard>
  </div>

  <UCard v-if="question" :ui="{ root: 'col-span-2' }">
    <QuestionMcq
      v-if="question.type === QuestionType.MCQ"
      :question="question.question"
    />
    <QuestionNonMcq v-else :question="question.question" />
  </UCard>

  <UCard :ui="{ root: 'col-span-2', body: 'flex' }">
    <UButton
      v-if="prevQuestionId"
      label="Previous"
      color="neutral"
      variant="outline"
      :to="{ query: { ...route.query, question: prevQuestionId } }"
    />
    <UButton
      v-if="nextQuestionId"
      label="Next"
      color="neutral"
      variant="outline"
      :to="{ query: { ...route.query, question: nextQuestionId } }"
      class="ml-auto"
    />
  </UCard>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionType } from '~/types'

definePageMeta({
  layout: 'paper',
})

const route = useRoute()
const paperId = parseInt(route.params.paperId as string)
const { data: paper } = await usePaper(paperId)
const { data: groupedQuestions } = await usePaperQuestions(paperId)
const { data: sortedCategories } = await usePaperCategories(paperId)

const lastVisitedQuestionForCategory = ref<Record<number, string>>({})

watch(
  route,
  newRoute => {
    const query = newRoute.query
    if (isNullOrUndefined(query) || isNullOrUndefined(query.category)) return
    const categoryId = parseInt(query.category as string)
    lastVisitedQuestionForCategory.value[categoryId] = query.question as string
  },
  { immediate: true }
)

function getQuestionIdForCategoryId(categoryId: number) {
  const categoryQuestions = groupedQuestions.value?.[categoryId]
  if (isNullOrUndefined(categoryQuestions)) return
  const questionId =
    lastVisitedQuestionForCategory.value[categoryId] ?? categoryQuestions[0].id
  return questionId
}

// Add initial `category` and `question` queries, if missing
if (!route.query.category && sortedCategories.value?.length) {
  const categoryId = sortedCategories.value[0].id
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
  if (currentQuestionIdx.value === -1) return null
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
