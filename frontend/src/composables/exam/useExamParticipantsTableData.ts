// import { formatTimeAgo } from '@vueuse/core'
import { UBadge, UButton } from '#components'
import type { TableColumn } from '@nuxt/ui'
import {
  ExamParticipantStatus,
  type ExamParticipantResponse,
  type ExamPermission,
} from '~/types'

const participantStatusColors = {
  [ExamParticipantStatus.UNATTENDED]: 'error',
  [ExamParticipantStatus.INVITED]: 'info',
  [ExamParticipantStatus.STARTED]: 'success',
  [ExamParticipantStatus.ENDED]: 'warning',
  [ExamParticipantStatus.EVALUATED]: 'success',
} as const

const participantStatusText = {
  [ExamParticipantStatus.UNATTENDED]: 'Unattended',
  [ExamParticipantStatus.INVITED]: 'Invited',
  [ExamParticipantStatus.STARTED]: 'Started',
  [ExamParticipantStatus.ENDED]: 'Pending evaluation',
  [ExamParticipantStatus.EVALUATED]: 'Evaluated',
} as const

export async function useExamParticipantsTableData(
  examId: number,
  examPermission: Ref<ExamPermission | null>
) {
  const { data: participants } = await useExamParticipants(examId)

  // const now = new Date()

  // function showEndTime(dateString: string) {
  //   const date = new Date(dateString)
  //   return `${now > date ? 'Ended' : 'Ends'} ${formatTimeAgo(date)}`
  // }

  const columns: TableColumn<ExamParticipantResponse>[] = [
    {
      header: 'Name',
      cell: ({ row }) => {
        const participantId = row.original.id
        const name = row.original.first_name + ' ' + row.original.last_name

        if (
          !examPermission.value?.can_evaluate ||
          [
            ExamParticipantStatus.UNATTENDED,
            ExamParticipantStatus.INVITED,
            ExamParticipantStatus.STARTED,
          ].includes(row.original.status)
        ) {
          return name
        }

        return h(UButton, {
          label: name,
          to: `/exams/${examId}/evaluation/${participantId}`,
          variant: 'link',
          ui: { base: 'px-0' },
        })
      },
    },
    {
      accessorKey: 'email',
      header: 'Email',
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => {
        const statusValue = row.getValue<ExamParticipantStatus>('status')
        const color = participantStatusColors[statusValue]

        return h(
          UBadge,
          { class: 'capitalize', variant: 'subtle', color },
          () => participantStatusText[statusValue]
        )
      },
    },
  ]

  return {
    data: participants,
    columns,
  }
}
