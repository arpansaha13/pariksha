/**
 * Delete multiple exams by their IDs
 */
export async function deleteExams(examIds: ExamId[]): Promise<void> {
  if (examIds.length === 0) {
    console.warn('deleteExams called with empty argument.')
    return
  }

  const { $api } = useNuxtApp()
  const toast = useToast()
  const { data: exams } = useNuxtData<Exam[]>(AsyncDataKeys.EXAMS)

  if (!exams.value) return

  // Store exams for potential rollback
  const previousExams = [...exams.value]

  // Optimistically remove exams from the list
  exams.value = exams.value.filter(exam => !examIds.includes(exam.id))

  try {
    await $api('/api/exams', {
      method: 'DELETE',
      body: {
        exam_ids: examIds,
      },
    })

    // Refresh exams data to ensure consistency
    await refreshNuxtData(AsyncDataKeys.EXAMS)
  } catch {
    // Rollback on error
    exams.value = previousExams

    toast.add({
      id: ToastId.DELETE_EXAM_FAILED,
      title: `Failed to delete ${examIds.length === 1 ? 'exam' : 'exams'}.`,
      color: 'error',
    })
  }
}
