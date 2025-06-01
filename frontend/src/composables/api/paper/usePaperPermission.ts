export function usePaperPermission(paperId: PaperId) {
  const { $api } = useNuxtApp()

  return useAsyncData<PaperPermission>(
    UseAsyncDataKeys.paper_permission(paperId),
    () => $api(`/api/papers/${paperId}/permissions`)
  )
}
