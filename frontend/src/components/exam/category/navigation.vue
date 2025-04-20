<template>
  <!-- subtract button-width and gap -->
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
</template>

<script setup lang="ts">
import type { ExamCategory } from '~/types'

const props = defineProps({
  sortedCategories: {
    type: Array as PropType<ExamCategory[]>,
    required: true,
  },
  getQuestionIdForCategoryId: {
    type: Function as PropType<(categoryId: number) => string | undefined>,
    required: true,
  },
})

const categoryLinks = computed(() => {
  return props.sortedCategories.map(category => ({
    label: category.name,
    to: {
      query: {
        category: category.id,
        question: props.getQuestionIdForCategoryId(category.id),
      },
    },
    exactQuery: true,
    replace: true,
  }))
})
</script>
