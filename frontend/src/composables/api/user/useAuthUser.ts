export function useAuthUser() {
  const { $api } = useNuxtApp()

  return useAsyncData<User>(
    UseAsyncDataKeys.auth_user,
    () => $api(`/api/users/me`),
    {
      transform: (res: string | User) =>
        typeof res === 'string' ? JSON.parse(res) : res,
    }
  )
}
