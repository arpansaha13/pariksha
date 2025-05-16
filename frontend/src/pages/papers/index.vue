<template>
  <main>
    <UCard :ui="{ body: '!py-2' }">
      <template #header>
        <h1 class="heading">Papers</h1>
      </template>

      <UTable
        :data="papersData ?? undefined"
        :loading="papersPending"
        :columns="columns"
        class="flex-1"
      />
    </UCard>
  </main>
</template>

<script setup lang="ts">
import { ConfirmModal } from '#components'
import type { DropdownMenuItem, TableColumn } from '@nuxt/ui'
import type { Paper, PaperQuestionCounts } from '~/types'

const UButton = resolveComponent('UButton')
const UDropdownMenu = resolveComponent('UDropdownMenu')

const { data: papersData, pending: papersPending } = await usePapers()

const toast = useToast()
const newExamStore = useNewExamStore()
const { copy, isSupported } = useClipboard()

const overlay = useOverlay()
const confirmModal = overlay.create(ConfirmModal)

function createExamWithPaper(paper: Paper) {
  newExamStore.clear()
  newExamStore.paper_id = paper.id
  newExamStore.duration_hours = calcHours(paper.duration_minutes ?? 0)
  newExamStore.duration_minutes = calcRemainderMinutes(
    paper.duration_minutes ?? 0
  )
  return navigateTo(`/exams/new`)
}

async function handleDeletePaper(paper: Paper) {
  const instance = confirmModal.open({
    title: 'Delete paper',
    description: `Are you sure you want to to delete ${paper.title}? This action cannot be undone.`,
    confirmLabel: 'Delete paper',
  })
  const shouldDelete = await instance.result

  if (shouldDelete) {
    deletePapers([paper.id])
  }
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
      return formatDurationMinutes(durationMinutes)
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
      const items: DropdownMenuItem[][] = [
        [
          {
            label: 'Create exam',
            icon: 'i-lucide-bookmark-plus',
            onSelect() {
              createExamWithPaper(row.original)
            },
          },
        ],
        [
          {
            label: 'Delete',
            color: 'error',
            icon: 'i-heroicons-trash',
            onSelect() {
              handleDeletePaper(row.original)
            },
          },
        ],
      ]

      if (isSupported) {
        items[0].push({
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
</script>
