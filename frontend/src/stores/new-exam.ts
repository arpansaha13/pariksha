import { defineStore } from 'pinia'
import type { CalendarDateTime } from '@internationalized/date'

export const useNewExamStore = defineStore(newExamStoreId, {
  state: () => ({
    title: null as string | null,
    type: null as ExamAccessType | null,
    paper_id: null as PaperId | null,
    startDate: null as CalendarDateTime | null,
    endDate: null as CalendarDateTime | null,
    duration_hours: null as number | null,
    duration_minutes: null as number | null,
  }),
})
