export default defineNuxtRouteMiddleware((to, from) => {
  const previousPath = useState(UseStateKeys.PreviousPath)
  previousPath.value = from.fullPath === to.fullPath ? null : from.fullPath
})
