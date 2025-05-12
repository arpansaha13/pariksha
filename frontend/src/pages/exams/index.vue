<template>
  <main>
    <UCard :ui="{ body: '!py-2' }">
      <template #header>
        <h1 class="heading">Exams</h1>
      </template>

      <UTable
        :loading="examsPending"
        :data="examsData ?? undefined"
        :columns="columns"
      />
    </UCard>
  </main>
</template>

<script setup lang="ts">
import { ConfirmModal } from '#components'
import type { DropdownMenuItem, TableColumn } from '@nuxt/ui'
import { formatTimeAgo } from '@vueuse/core'
import type { Exam } from '~/types'

const UButton = resolveComponent('UButton')
const UDropdownMenu = resolveComponent('UDropdownMenu')

const { data: examsData, pending: examsPending } = await useExams()

const toast = useToast()
const { copy, isSupported } = useClipboard()

const overlay = useOverlay()
const confirmModal = overlay.create(ConfirmModal)

async function handleDeleteExam(exam: Exam) {
  const instance = confirmModal.open({
    title: 'Delete exam',
    description: `Are you sure you want to to delete ${exam.title}? This action cannot be undone.`,
    confirmLabel: 'Delete exam',
  })
  const shouldDelete = await instance.result

  if (shouldDelete) {
    deleteExams([exam.id])
  }
}

function getLinkToExam(examId: number) {
  return `/exams/${examId}`
}

const columns: TableColumn<Exam>[] = [
  {
    accessorKey: 'title',
    header: 'Title',
    cell: ({ row }) => {
      const examId = row.original.id
      const title = row.getValue<string>('title')

      return h(UButton, {
        label: title,
        to: getLinkToExam(examId),
        variant: 'link',
        ui: { base: 'px-0' },
      })
    },
  },
  {
    accessorKey: 'starts_at',
    header: 'Starts/Started at',
    cell: ({ row }) => {
      const startsAt = row.getValue<string>('starts_at')
      return formatTimeAgo(new Date(startsAt))
    },
  },
  {
    accessorKey: 'ends_at',
    header: 'Ends/Ended at',
    cell: ({ row }) => {
      const endsAt = row.getValue<string>('ends_at')
      return formatTimeAgo(new Date(endsAt))
    },
  },
  {
    id: 'actions',
    enableHiding: false,
    cell: ({ row }) => {
      const items: DropdownMenuItem[][] = []

      if (isSupported) {
        items.push([
          {
            label: 'Copy link',
            icon: 'i-lucide-link',
            onSelect() {
              copy(getLinkToExam(row.original.id))

              toast.add({
                id: ToastId.COPIED_TO_CLIPBOARD,
                title: 'Exam link copied!',
                color: 'success',
                icon: 'i-lucide-clipboard-copy',
              })
            },
          },
        ])
      }

      items.push([
        {
          label: 'Delete',
          color: 'error',
          icon: 'i-heroicons-trash',
          onSelect() {
            handleDeleteExam(row.original)
          },
        },
      ])

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
</script>
