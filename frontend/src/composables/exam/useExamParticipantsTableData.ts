// import { formatTimeAgo } from '@vueuse/core'

import { UBadge } from '#components'
import type { TableColumn } from '@nuxt/ui'
import { ExamParticipantStatus, type ExamParticipantResponse } from '~/types'

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

export async function useExamParticipantsTableData(examId: number) {
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
        const name = row.original.first_name + ' ' + row.original.last_name
        return name
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
        const statusValue = row.getValue('status') as ExamParticipantStatus
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
