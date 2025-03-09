import type { Paper } from '~/types/models'

export async function createPaper() {
  const res = await $fetch<string>('/api/papers', {
    method: 'POST',
    ...getFetchOptions(),
  })

  return JSON.parse(res) as Paper
}
