export function usePapers() {
  const { $api } = useNuxtApp()

  return useAsyncData<Paper[]>(AsyncDataKeys.PAPERS, () => $api('/api/papers'))
}
