export function usePaperPermission(paperId: PaperId) {
  const { $api } = useNuxtApp()

  return useAsyncData<PaperPermission>(
    AsyncDataKeys.PAPER_PERMISSION(paperId),
    () => $api(`/api/papers/${paperId}/permissions`)
  )
}
