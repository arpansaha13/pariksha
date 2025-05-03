import type { User } from '../../../types/user'

export interface UpdateUserPayload {
  username?: string
  first_name?: string
  last_name?: string
}

export async function updateAuthUser(body: UpdateUserPayload) {
  await $fetch<User>(`/api/users/me`, {
    method: 'PATCH',
    body,
    ...getFetchOptions(),
  })
  return refreshNuxtData(AsyncDataKeys.AUTH_USER)
}
