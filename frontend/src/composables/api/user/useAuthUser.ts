import type { User } from '~/types/user'

export function useAuthUser() {
  const { $api } = useNuxtApp()

  return useAsyncData<User>(
    AsyncDataKeys.AUTH_USER,
    () => $api(`/api/users/me`),
    {
      transform: (res: string | User) =>
        typeof res === 'string' ? JSON.parse(res) : res,
    }
  )
}
