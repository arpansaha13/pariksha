import { defineStore } from 'pinia'
import type { CalendarDateTime } from '@internationalized/date'
import type { ExamAccessType } from '~/types/exam'

export const useNewExamStore = defineStore('new-exam', {
  state: () => ({
    title: null as string | null,
    type: null as ExamAccessType | null,
    paper_id: null as number | null,
    startDate: null as CalendarDateTime | null,
    endDate: null as CalendarDateTime | null,
  }),
  actions: {
    clear() {
      this.title = null
      this.type = null
      this.paper_id = null
      this.startDate = null
      this.endDate = null
    },
  },
})
