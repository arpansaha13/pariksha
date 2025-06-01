import { defineStore } from 'pinia'

export const useAuthStore = defineStore(authStoreId, {
  state: () => ({
    signUpEmail: null as string | null,
    forgotPassEmail: null as string | null,
  }),
  actions: {
    setSignUpEmail(email: string) {
      this.signUpEmail = email
    },
    clearSignUpEmail() {
      this.signUpEmail = null
    },
    setForgotPassEmail(email: string) {
      this.forgotPassEmail = email
    },
    clearForgotPassEmail() {
      this.forgotPassEmail = null
    },
  },
  persist: {
    key: 'authStore',
    storage: piniaPluginPersistedstate.sessionStorage(),
  },
})
