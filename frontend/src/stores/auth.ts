import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    signUpEmail: null as string | null,
  }),
  actions: {
    setSignUpEmail(email: string) {
      this.signUpEmail = email
    },
    clearSignUpEmail() {
      this.signUpEmail = null
    },
  },
  persist: {
    key: 'signUpEmail',
    storage: piniaPluginPersistedstate.sessionStorage(),
  },
})
