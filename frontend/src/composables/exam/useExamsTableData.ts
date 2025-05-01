import { formatTimeAgo } from '@vueuse/core'
import type { TableColumn } from '@nuxt/ui'
import { UButton } from '#components'
import type { Exam } from '~/types'

// interface ExamTableData {
//   id: number
//   title: string
//   starts_at: string
//   ends_at: string
// }

export async function useExamsTableData() {
  const { data: exams } = await useExams()

  const columns: TableColumn<Exam>[] = [
    {
      accessorKey: 'title',
      header: 'Title',
      cell: ({ row }) => {
        const examId = row.original.id
        const title = row.getValue('title') as string

        return h(
          UButton,
          {
            to: `/exams/${examId}`,
            variant: 'link',
            ui: { base: 'px-0' },
          },
          title
        )
      },
    },
    {
      accessorKey: 'starts_at',
      header: 'Starts/Started at',
      cell: ({ row }) => {
        const startsAt = row.getValue('starts_at') as string
        return formatTimeAgo(new Date(startsAt))
      },
    },
    {
      accessorKey: 'ends_at',
      header: 'Ends/Ended at',
      cell: ({ row }) => {
        const endsAt = row.getValue('ends_at') as string
        return formatTimeAgo(new Date(endsAt))
      },
    },
  ]

  return {
    data: exams,
    columns,
  }
}
