import type { CalendarDateTime } from '@internationalized/date'
import { defineStore } from 'pinia'

export const useNewExamStore = defineStore('new-exam', {
  state: () => ({
    title: null as string | null,
    type: null as string | null,
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
