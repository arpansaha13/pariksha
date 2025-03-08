import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import ForgotPasswordPage from './forgot-password.vue'

const { mockNavigateTo, mockForgotPassword, mockToast, mockAuthStore } =
  vi.hoisted(() => ({
    mockNavigateTo: vi.fn(),
    mockForgotPassword: vi.fn(),
    mockToast: {
      add: vi.fn(),
    },
    mockAuthStore: {
      setForgotPassEmail: vi.fn(),
    },
  }))

mockNuxtImport('navigateTo', () => mockNavigateTo)
mockNuxtImport('forgotPassword', () => mockForgotPassword)
mockNuxtImport('useToast', () => () => mockToast)
mockNuxtImport('useAuthStore', () => () => mockAuthStore)

describe('Forgot Password Page', () => {
  let component: Awaited<
    ReturnType<typeof mountSuspended<typeof ForgotPasswordPage>>
  >

  beforeEach(async () => {
    component = await mountSuspended(ForgotPasswordPage)
    vi.clearAllMocks()
  })

  describe('Page Layout', () => {
    it('renders the forgot password form', () => {
      expect(component.find('form').exists()).toBe(true)
      expect(component.find('input[name="email"]').exists()).toBe(true)
      expect(component.find('button[type="submit"]').exists()).toBe(true)
    })

    it('displays login link', () => {
      const link = component.find('a[href="/auth/login"]')
      expect(link.exists()).toBe(true)
      expect(link.text()).toContain('Login')
    })
  })

  describe('Form Functionality', () => {
    it('updates email value on input', async () => {
      const email = 'test@example.com'
      const emailInput = component.find('input[name="email"]')
      await emailInput.setValue(email)

      // @ts-expect-error vm does not detect ts types
      expect(component.vm.forgotPasswordFormData._value.email).toBe(email)
    })

    it('requires email field', () => {
      const emailInput = component.find('input[name="email"]')
      expect(emailInput.attributes('required')).toBeDefined()
    })
  })

  describe('Form Submission', () => {
    const testEmail = 'test@example.com'

    beforeEach(async () => {
      const emailInput = component.find('input[name="email"]')
      await emailInput.setValue(testEmail)
    })

    it('calls forgot password API with correct email', async () => {
      const form = component.find('form')
      await form.trigger('submit')

      expect(mockForgotPassword).toHaveBeenCalledWith({ email: testEmail })
    })

    it.skip('stores email in auth store on success', async () => {
      mockForgotPassword.mockResolvedValueOnce({})

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockAuthStore.setForgotPassEmail).toHaveBeenCalledWith(testEmail)
    })

    it.skip('navigates to reset password page on success', async () => {
      mockForgotPassword.mockResolvedValueOnce({})

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockNavigateTo).toHaveBeenCalledWith('/auth/reset-password')
    })

    it.skip('shows error toast when email not found', async () => {
      const error = new Error()
      Object.defineProperty(error, 'status', { value: 404 })
      mockForgotPassword.mockRejectedValueOnce(error)

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockToast.add).toHaveBeenCalledWith({
        id: ToastId.FORGOT_PASSWORD_FAILED,
        color: 'red',
        title: 'Failed to send OTP',
        description: 'This email is not registered.',
        icon: 'i-heroicons-exclamation-circle',
      })
    })
  })
})
