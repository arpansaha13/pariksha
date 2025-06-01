export function usePaper(paperId: PaperId) {
  const { $api } = useNuxtApp()

  return useAsyncData<Paper>(AsyncDataKeys.PAPERS_PAPER(paperId), () =>
    $api(`/api/papers/${paperId}`)
  )
}
