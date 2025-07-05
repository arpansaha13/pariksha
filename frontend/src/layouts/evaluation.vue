<template>
  <div class="flex h-screen flex-col">
    <UContainer
      as="header"
      class="bg-default flex justify-between py-3 shadow-sm sm:col-span-full sm:bg-transparent sm:shadow-none"
    >
      <div class="flex items-center gap-1.5">
        <UButton
          :to="`/exams/${examId}`"
          icon="heroicons:arrow-left-16-solid"
          color="neutral"
          variant="link"
          class="sm:hidden"
        />

        <slot name="title" />
      </div>

      <div class="flex items-center gap-1.5 sm:gap-2">
        <UButton
          :to="`/exams/${examId}`"
          icon="i-heroicons-arrow-uturn-left"
          color="neutral"
          variant="outline"
          class="hidden sm:flex"
        />

        <slot name="submit" />

        <USlideover
          side="right"
          :title="questionPalletTitle"
          :ui="{ body: 'px-0! pt-0!' }"
        >
          <UButton
            icon="lucide:file-question-mark"
            size="sm"
            color="neutral"
            variant="outline"
            class="sm:hidden"
          />

          <template #body>
            <!-- Mobile Category Nav -->
            <div :class="categoryNavBorder">
              <slot name="category-nav" />
            </div>

            <!-- Mobile Question Nav -->
            <div class="p-4">
              <slot name="question-nav" />
            </div>
          </template>
        </USlideover>
      </div>
    </UContainer>

    <UContainer
      as="main"
      class="mt-2.5 grid grow grid-cols-1 grid-rows-[1fr_auto] gap-y-2.5 pb-4 sm:mt-0 sm:grid-cols-3 sm:grid-rows-[auto_1fr_auto] sm:gap-x-6 sm:gap-y-4"
    >
      <!-- Desktop Category Nav -->
      <div
        :class="[
          categoryNavBorder,
          'hidden sm:col-span-2 sm:flex sm:items-center sm:justify-between sm:gap-x-2',
        ]"
      >
        <slot name="category-nav" />
      </div>

      <!-- Mobile Question Nav -->
      <UCard
        :ui="{
          root: 'col-start-3 row-span-2 row-start-1 overflow-hidden flex-col hidden sm:flex',
          body: 'overflow-auto grow',
        }"
      >
        <template #header>
          <h2 class="text-lg font-semibold">{{ questionPalletTitle }}</h2>
        </template>

        <div>
          <slot name="question-nav" />
        </div>
      </UCard>

      <div
        class="-m-[2px] flex flex-col gap-y-2.5 overflow-auto p-[2px] sm:col-span-2 sm:row-span-2"
      >
        <slot name="body" />
      </div>

      <UCard
        v-if="$slots.footer"
        :ui="{ root: 'sm:col-start-3', body: 'flex' }"
      >
        <slot name="footer" />
      </UCard>
    </UContainer>
  </div>
</template>

<script setup lang="ts">
const categoryNavBorder = 'border-b border-gray-200 dark:border-gray-800'
const questionPalletTitle = 'Question Pallet'

const route = useRoute()
const examId = route.params.examId as ExamId
</script>
