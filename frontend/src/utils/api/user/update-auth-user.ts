interface UpdateUserPayload {
  username?: string
  first_name?: string
  last_name?: string
}

export async function updateAuthUser(body: UpdateUserPayload) {
  const { $api } = useNuxtApp()

  await $api<User>(`/api/users/me`, {
    method: 'PATCH',
    body,
  })

  return refreshNuxtData(AsyncDataKeys.AUTH_USER)
}
