<template>
  <main>
    <h1 class="heading mb-6">Papers</h1>

    <UCard v-if="papers !== null" :ui="{ body: 'p-0 sm:p-0' }">
      <ul class="divide-y divide-gray-200">
        <li
          v-for="paper in papers"
          :key="paper.id"
          class="group grid grid-cols-[1fr_1fr_auto] items-center px-4 py-3"
        >
          <div>
            <h2 class="text-sm font-medium">{{ paper.title }}</h2>
          </div>

          <div class="text-sm text-gray-500">
            <p>{{ countTotalQuestions(paper.question_counts) }} questions</p>
            <p>{{ paper.duration_minutes ?? 0 }} minutes</p>
          </div>

          <div class="invisible space-x-1.5 group-hover:visible">
            <UTooltip text="Open">
              <UButton
                :to="`/papers/${paper.id}`"
                icon="i-heroicons-arrow-right-end-on-rectangle"
                size="sm"
                color="neutral"
                square
                variant="outline"
                no-prefetch
              />
            </UTooltip>

            <UTooltip text="Create exam">
              <UButton
                icon="i-lucide-bookmark-plus"
                size="sm"
                color="neutral"
                square
                variant="outline"
                @click="createExamWithPaper(paper)"
              />
            </UTooltip>
          </div>
        </li>
      </ul>
    </UCard>
  </main>
</template>

<script setup lang="ts">
import type { Paper } from '~/types'

const { data: papers } = await usePapers()
const newExamStore = useNewExamStore()

function createExamWithPaper(paper: Paper) {
  newExamStore.clear()
  newExamStore.paper_id = paper.id
  newExamStore.duration_hours = calcHours(paper.duration_minutes ?? 0)
  newExamStore.duration_minutes = calcRemainderMinutes(
    paper.duration_minutes ?? 0
  )
  return navigateTo(`/exams/new`)
}
</script>
