<template>
  <main>
    <h1 class="heading mb-6">Exams</h1>

    <UCard
      v-if="exams !== null"
      :ui="{
        body: 'p-0 sm:p-0',
      }"
    >
      <ul class="divide-y divide-gray-200">
        <li
          v-for="exam in exams"
          :key="exam.id"
          class="group grid grid-cols-3 items-center justify-between gap-4 px-4 py-3"
        >
          <div>
            <h2 class="text-sm">
              <UButton :to="`/exams/${exam.id}`" variant="link">
                {{ exam.title }}
              </UButton>
            </h2>
          </div>

          <div>
            <p class="text-sm text-gray-500">
              {{ showStartTime(exam.starts_at) }}
            </p>
          </div>

          <div>
            <p class="text-sm text-gray-500">
              {{ showEndTime(exam.starts_at) }}
            </p>
          </div>
        </li>
      </ul>
    </UCard>
  </main>
</template>

<script setup lang="ts">
import { formatTimeAgo } from '@vueuse/core'

const { data: exams } = await useExams()
const now = new Date()

function showStartTime(dateString: string) {
  const date = new Date(dateString)
  return `${now > date ? 'Started' : 'Starts'} ${formatTimeAgo(date)}`
}

function showEndTime(dateString: string) {
  const date = new Date(dateString)
  return `${now > date ? 'Ended' : 'Ends'} ${formatTimeAgo(date)}`
}
</script>
