export function usePapers() {
  const { $api } = useNuxtApp()

  return useAsyncData<Paper[]>(UseAsyncDataKeys.papers, () =>
    $api('/api/papers')
  )
}
