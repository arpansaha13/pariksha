<template>
  <!-- subtract button-width and gap -->
  <ScrollAreaRoot class="max-w-[calc(100%-44px)]">
    <ScrollAreaViewport>
      <UNavigationMenu
        :items="categoryLinks"
        color="primary"
        orientation="horizontal"
        variant="link"
        highlight
      >
        <template #item="{ item }">
          <UChip
            :show="!!unsavedCount[item.to.query.category]"
            :ui="{
              base: '-top-1 -right-1.5',
            }"
          >
            <span class="truncate">{{ item.label }}</span>
          </UChip>
        </template>
      </UNavigationMenu>
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
import type { QuestionCategory, QuestionId } from '~/types'

const props = defineProps({
  sortedCategories: {
    type: Array as PropType<QuestionCategory[]>,
    required: true,
  },
  unsavedCount: {
    type: Object as PropType<Record<number, number>>,
    required: true,
  },
  getQuestionIdForCategoryId: {
    type: Function as PropType<(categoryId: number) => string | QuestionId.ADD>,
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
  }))
})
</script>
