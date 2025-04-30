// import { formatTimeAgo } from '@vueuse/core'

import { UBadge } from '#components'
import type { TableColumn } from '@nuxt/ui'
import { ExamParticipantStatus } from '~/types'

interface ExamParticipantTableData {
  name: string
  email: string
  status: ExamParticipantStatus
}

const participantStatusColors = {
  [ExamParticipantStatus.UNATTENDED]: 'error',
  [ExamParticipantStatus.INVITED]: 'info',
  [ExamParticipantStatus.STARTED]: 'warning',
  [ExamParticipantStatus.ENDED]: 'success',
  [ExamParticipantStatus.EVALUATED]: 'neutral',
} as const

const participantStatusText = {
  [ExamParticipantStatus.UNATTENDED]: 'Unattended',
  [ExamParticipantStatus.INVITED]: 'Invited',
  [ExamParticipantStatus.STARTED]: 'Started',
  [ExamParticipantStatus.ENDED]: 'Ended',
  [ExamParticipantStatus.EVALUATED]: 'Evaluated',
} as const

export async function useExamParticipantsTableData(examId: number) {
  const { data: participants } = await useExamParticipants(examId)

  // const now = new Date()

  // function showEndTime(dateString: string) {
  //   const date = new Date(dateString)
  //   return `${now > date ? 'Ended' : 'Ends'} ${formatTimeAgo(date)}`
  // }

  const data = computed(() => {
    if (!participants.value) return null
    return participants.value.map(participant => ({
      name: participant.first_name + ' ' + participant.last_name,
      email: participant.email,
      status: participant.status,
      // started_at: showEndTime(participant.started_at),
      // ended_at: showEndTime(participant.ended_at),
    }))
  })

  const columns: TableColumn<ExamParticipantTableData>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
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
    data,
    columns,
  }
}
