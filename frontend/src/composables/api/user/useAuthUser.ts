import type { User } from '~/types/user'

export function useAuthUser() {
  const fetchOptions = getFetchOptions()

  return useAsyncData<User>(
    AsyncDataKeys.AUTH_USER,
    () => $fetch(`/api/users/me`, fetchOptions),
    {
      transform: (res: string | User) =>
        typeof res === 'string' ? JSON.parse(res) : res,
    }
  )
}
