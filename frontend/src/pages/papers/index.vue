<template>
  <main>
    <h1 class="mb-6 text-3xl font-bold">Papers</h1>

    <UCard
      v-if="papers !== null"
      :ui="{
        body: {
          padding: 'p-0 sm:p-0',
        },
      }"
    >
      <ul class="divide-y">
        <li
          v-for="paper in papers"
          :key="paper.id"
          class="group flex items-center justify-between px-4 py-3 hover:bg-gray-100"
        >
          <div>
            <h2 class="text-sm font-medium">{{ paper.title }}</h2>
          </div>
          <div class="invisible space-x-1.5 group-hover:visible">
            <UTooltip text="Open">
              <UButton
                :to="`/papers/${paper.id}`"
                icon="i-heroicons-arrow-right-end-on-rectangle"
                size="xs"
                color="white"
                square
                variant="solid"
                no-prefetch
              />
            </UTooltip>
            <UTooltip text="Edit">
              <UButton
                :to="`/papers/${paper.id}/edit`"
                icon="i-heroicons-pencil-square"
                size="xs"
                color="white"
                square
                variant="solid"
                no-prefetch
              />
            </UTooltip>
          </div>
        </li>
      </ul>
    </UCard>
  </main>
</template>

<script setup lang="ts">
const { data: papers } = await usePapers()
</script>
