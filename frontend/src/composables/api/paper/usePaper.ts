export function usePaper(paperId: PaperId) {
  const { $api } = useNuxtApp()

  return useAsyncData<Paper>(UseAsyncDataKeys.paper(paperId), () =>
    $api(`/api/papers/${paperId}`)
  )
}
