export function usePaperPermission(paperId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<PaperPermission>(
    AsyncDataKeys.PAPER_PERMISSION(paperId),
    () => $api(`/api/papers/${paperId}/permissions`)
  )
}
