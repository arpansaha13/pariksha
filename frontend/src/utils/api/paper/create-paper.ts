export async function createPaper() {
  const { $api } = useNuxtApp()

  const res = await $api<string>('/api/papers', {
    method: 'POST',
  })

  return JSON.parse(res) as Paper
}
