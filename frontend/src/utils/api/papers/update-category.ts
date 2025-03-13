interface UpdateCategoryBody {
  name: string
}

export async function updateCategory(
  categoryId: number,
  body: UpdateCategoryBody
): Promise<void> {
  return $fetch(`/api/categories/${categoryId}`, {
    method: 'POST',
    body,
    ...getFetchOptions(),
  })
}
