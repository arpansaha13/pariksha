<template>
  <main>
    <UCard :ui="{ body: '!py-2' }">
      <template #header>
        <h1 class="heading">Papers</h1>
      </template>

      <UTable
        v-if="!isNullOrUndefined(papers)"
        :data="papers"
        :columns="columns"
        class="flex-1"
      />
    </UCard>
  </main>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'

const UButton = resolveComponent('UButton')
const UDropdownMenu = resolveComponent('UDropdownMenu')

const { data: papers } = await usePapers()

const toast = useToast()
const newExamStore = useNewExamStore()
const { copy, isSupported } = useClipboard()

function createExamWithPaper(paper: Paper) {
  newExamStore.clear()
  newExamStore.paper_id = paper.id
  newExamStore.duration_hours = calcHours(paper.duration_minutes ?? 0)
  newExamStore.duration_minutes = calcRemainderMinutes(
    paper.duration_minutes ?? 0
  )
  return navigateTo(`/exams/new`)
}

const columns: TableColumn<Paper>[] = [
  {
    accessorKey: 'title',
    header: 'Title',
    cell: ({ row }) => {
      const title = row.getValue('title') as string

      return h(UButton, {
        label: title,
        variant: 'link',
        to: getLinkToPaper(row.original.id),
        ui: { base: 'px-0' },
      })
    },
  },
  {
    accessorKey: 'question_counts',
    header: 'No. of questions',
    cell: ({ row }) => {
      const questionCounts = row.getValue(
        'question_counts'
      ) as PaperQuestionCounts
      return getQuestionCountsText(questionCounts)
    },
  },
  {
    accessorKey: 'duration_minutes',
    header: 'Duration',
    cell: ({ row }) => {
      const durationMinutes = row.getValue('duration_minutes') as number
      return getDurationMinutesText(durationMinutes)
    },
  },
  {
    accessorKey: 'max_score',
    header: 'Max score',
    cell: ({ row }) => {
      const maxScore = row.getValue('max_score') as number
      return maxScore ?? 0
    },
  },
  {
    id: 'actions',
    enableHiding: false,
    cell: ({ row }) => {
      const items = [
        {
          label: 'Create exam',
          icon: 'i-lucide-bookmark-plus',
          onSelect() {
            createExamWithPaper(row.original)
          },
        },
      ]

      if (isSupported) {
        items.push({
          label: 'Copy link',
          icon: 'i-lucide-link',
          onSelect() {
            copy(getLinkToPaper(row.original.id))

            toast.add({
              id: ToastId.COPIED_TO_CLIPBOARD,
              title: 'Paper link copied!',
              color: 'success',
              icon: 'i-lucide-clipboard-copy',
            })
          },
        })
      }

      return h(
        'div',
        { class: 'text-right' },
        h(
          UDropdownMenu,
          {
            'content': {
              align: 'end',
            },
            items,
            'aria-label': 'Actions dropdown',
          },
          () =>
            h(UButton, {
              'icon': 'i-lucide-ellipsis-vertical',
              'color': 'neutral',
              'variant': 'ghost',
              'class': 'ml-auto',
              'aria-label': 'Actions dropdown',
            })
        )
      )
    },
  },
]

function getLinkToPaper(paperId: number) {
  return `/papers/${paperId}`
}

function getQuestionCountsText(questionCounts: PaperQuestionCounts) {
  const totalCount = countTotalQuestions(questionCounts)

  if (totalCount === 1) return `${totalCount} question`
  return `${totalCount} questions`
}

function getDurationMinutesText(durationMinutes: number) {
  if (isNullOrUndefined(durationMinutes)) return '0 minutes'

  const hours = calcHours(durationMinutes)
  const remainingMinutes = calcRemainderMinutes(durationMinutes)

  if (hours === 0) return `${remainingMinutes} minutes`
  if (remainingMinutes === 0)
    return `${hours} ${hours === 1 ? 'hour' : 'hours'}`
  if (hours === 1)
    return `${hours} hour ${remainingMinutes} ${remainingMinutes === 1 ? 'minute' : 'minutes'}`
  return `${hours} hours  ${remainingMinutes} minutes`
}
</script>
