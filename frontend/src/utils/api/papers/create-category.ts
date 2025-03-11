export async function createCategory(paperId: number): Promise<void> {
  return $fetch(`/api/papers/${paperId}/categories`, {
    method: 'POST',
    ...getFetchOptions(),
  })
}
