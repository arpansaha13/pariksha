import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import VerificationPage from './verification.vue'

const {
  mockNavigateTo,
  mockVerifySignUpEmail,
  mockToast,
  mockAuthStore,
  mockRouter,
} = vi.hoisted(() => ({
  mockNavigateTo: vi.fn(),
  mockVerifySignUpEmail: vi.fn(),
  mockToast: {
    add: vi.fn(),
  },
  mockRouter: {
    replace: vi.fn(),
    resolve: vi.fn(),
  },
  mockAuthStore: {
    signUpEmail: 'test@example.com' as string | null,
    clearSignUpEmail: vi.fn(),
  },
}))

mockNuxtImport('navigateTo', () => mockNavigateTo)
mockNuxtImport('verifySignUpEmail', () => mockVerifySignUpEmail)
mockNuxtImport('useRouter', () => () => mockToast)
mockNuxtImport('useAuthStore', () => () => mockAuthStore)
mockNuxtImport('useRouter', () => () => mockRouter)

describe('Verification Page', () => {
  let component: Awaited<
    ReturnType<typeof mountSuspended<typeof VerificationPage>>
  >

  beforeEach(async () => {
    component = await mountSuspended(VerificationPage)
    vi.clearAllMocks()
  })

  describe('Page Layout', () => {
    it('renders the verification form', () => {
      expect(component.find('form').exists()).toBe(true)
      expect(component.find('input[inputmode="numeric"]').exists()).toBe(true)
      expect(component.find('button[type="submit"]').exists()).toBe(true)
    })

    it('displays stored email', () => {
      expect(component.text()).toContain('test@example.com')
    })
  })

  describe('Form Submission', () => {
    const testOtp = '123456'

    beforeEach(async () => {
      const otpInput = component.find('input[inputmode="numeric"]')
      await otpInput.setValue(testOtp)
    })

    it('calls verify API with correct data', async () => {
      const form = component.find('form')
      await form.trigger('submit')

      expect(mockVerifySignUpEmail).toHaveBeenCalledWith({
        email: mockAuthStore.signUpEmail,
        otp: testOtp,
      })
    })

    it.skip('navigates to home on success', async () => {
      mockVerifySignUpEmail.mockResolvedValueOnce({})

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockNavigateTo).toHaveBeenCalledWith('/')
    })

    it.skip('shows error toast on invalid OTP', async () => {
      const error = new Error()
      Object.defineProperty(error, 'status', { value: 401 })
      mockVerifySignUpEmail.mockRejectedValueOnce(error)

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockToast.add).toHaveBeenCalledWith({
        id: ToastId.VERIFY_SIGNUP_FAILED,
        color: 'red',
        title: 'Failed to verify email',
        description: 'Invalid OTP. Please try again.',
        icon: 'i-heroicons-exclamation-circle',
      })
    })

    it.skip('clears email from store after submission', async () => {
      const form = component.find('form')
      await form.trigger('submit')

      expect(mockAuthStore.clearSignUpEmail).toHaveBeenCalled()
    })
  })

  describe('Route Guard', () => {
    it('redirects to signup if email not set', async () => {
      mockAuthStore.signUpEmail = null
      await mountSuspended(VerificationPage)
      expect(mockRouter.replace).toHaveBeenCalledWith('/auth/signup')
    })
  })
})
