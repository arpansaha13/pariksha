export enum ExamAccessType {
  LINK = 'LINK',
  INVITE = 'INVITE',
}

export interface Exam {
  id: number
  title: string
  starts_at: string
  ends_at: string
  created_by: number
  type: ExamAccessType
  max_candidates_count: number
  paper_id: number
}
