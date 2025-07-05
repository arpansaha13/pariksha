<template>
  <div class="flex h-screen flex-col">
    <UContainer
      as="header"
      class="bg-default flex justify-between py-3 shadow-sm sm:col-span-full sm:bg-transparent sm:shadow-none"
    >
      <div class="flex items-center gap-1.5">
        <Icon
          name="i-heroicons-document-text"
          size="2rem"
          class="hidden sm:block"
        />

        <UButton
          to="/papers"
          icon="heroicons:arrow-left-16-solid"
          color="neutral"
          variant="link"
          class="sm:hidden"
        />

        <slot name="title" />
      </div>

      <div class="flex items-center gap-1.5 sm:gap-2">
        <UButton
          to="/papers"
          icon="i-heroicons-arrow-uturn-left"
          size="sm"
          color="neutral"
          variant="outline"
          class="hidden sm:flex"
        />

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
            <slot name="question-nav" />
          </template>

          <template #footer>
            <slot name="new-question-link" />
          </template>
        </USlideover>

        <slot name="header-trailing" />
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

      <!-- Desktop Question Nav -->
      <UCard
        :ui="{
          root: 'col-start-3 row-span-2 row-start-1 overflow-hidden flex-col hidden sm:flex',
          body: 'overflow-auto grow p-0 sm:p-0',
        }"
      >
        <template #header>
          <h2 class="text-lg font-semibold">{{ questionPalletTitle }}</h2>
        </template>

        <div>
          <slot name="question-nav" />
        </div>

        <template #footer>
          <slot name="new-question-link" />
        </template>
      </UCard>

      <UCard
        :ui="{
          root: 'sm:col-span-2 overflow-hidden',
          body: 'h-full overflow-auto',
        }"
      >
        <slot name="body" />
      </UCard>

      <UCard :ui="{ root: 'sm:col-span-2', body: 'flex justify-between' }">
        <div>
          <slot name="footer-leading" />
        </div>
        <div class="space-x-2">
          <slot name="footer-trailing" />
        </div>
      </UCard>
    </UContainer>
  </div>
</template>

<script setup lang="ts">
const categoryNavBorder = 'border-b border-gray-200 dark:border-gray-800'
const questionPalletTitle = 'Question Pallet'
</script>
