export function logout() {
  const { $api } = useNuxtApp()

  return $api('/api/logout', {
    method: 'POST',
  })
}
