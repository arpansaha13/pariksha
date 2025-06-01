export async function deletePapers(paperIds: PaperId[]): Promise<void> {
  if (paperIds.length === 0) {
    console.warn('deletePaper called with empty argument.')
    return
  }

  const { $api } = useNuxtApp()
  const toast = useToast()
  const { data: papers } = useNuxtData<Paper[]>(UseAsyncDataKeys.papers)

  if (!papers.value) return

  // Store papers for potential rollback
  const previousPapers = [...papers.value]

  // Optimistically remove papers from the list
  papers.value = papers.value.filter(paper => !paperIds.includes(paper.id))

  try {
    await $api('/api/papers', {
      method: 'DELETE',
      body: {
        paper_ids: paperIds,
      },
    })

    // Refresh papers data to ensure consistency
    await refreshNuxtData(UseAsyncDataKeys.papers)
  } catch {
    // Rollback on error
    papers.value = previousPapers

    toast.add({
      id: ToastId.DELETE_PAPER_FAILED,
      title: `Failed to delete ${paperIds.length === 1 ? 'paper' : 'papers'}.`,
      color: 'error',
    })
  }
}
