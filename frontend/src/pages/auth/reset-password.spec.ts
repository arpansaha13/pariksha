import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import ResetPasswordPage from './reset-password.vue'

const {
  mockNavigateTo,
  mockResetPassword,
  mockToast,
  mockAuthStore,
  mockRouter,
} = vi.hoisted(() => ({
  mockNavigateTo: vi.fn(),
  mockResetPassword: vi.fn(),
  mockRouter: {
    replace: vi.fn(),
    resolve: vi.fn(),
  },
  mockToast: {
    add: vi.fn(),
  },
  mockAuthStore: {
    forgotPassEmail: 'test@example.com' as string | null,
    setForgotPassEmail: vi.fn(),
  },
}))

mockNuxtImport('navigateTo', () => mockNavigateTo)
mockNuxtImport('resetPassword', () => mockResetPassword)
mockNuxtImport('useToast', () => () => mockToast)
mockNuxtImport('useAuthStore', () => () => mockAuthStore)
mockNuxtImport('useRouter', () => () => mockRouter)

describe('Reset Password Page', () => {
  let component: Awaited<
    ReturnType<typeof mountSuspended<typeof ResetPasswordPage>>
  >

  beforeEach(async () => {
    component = await mountSuspended(ResetPasswordPage)
    vi.clearAllMocks()
  })

  describe('Page Layout', () => {
    it('renders the reset password form', () => {
      expect(component.find('form').exists()).toBe(true)
      expect(component.find('input[name="otp"]').exists()).toBe(true)
      expect(component.find('input[name="newPassword"]').exists()).toBe(true)
      expect(component.find('input[name="confirmNewPassword"]').exists()).toBe(
        true
      )
      expect(component.find('button[type="submit"]').exists()).toBe(true)
    })

    it('displays stored email', () => {
      expect(component.text()).toContain('test@example.com')
    })
  })

  describe('Form Validation', () => {
    it('shows error when passwords do not match', async () => {
      const passwordInput = component.find('input[name="newPassword"]')
      await passwordInput.setValue('password123')

      const confirmPasswordInput = component.find(
        'input[name="confirmNewPassword"]'
      )
      await confirmPasswordInput.setValue('password456')

      // @ts-expect-error vm does not detect ts types
      const errors = component.vm.validate(
        // @ts-expect-error vm does not detect ts types
        component.vm.resetPasswordFormData._value
      )
      expect(errors).toContainEqual({
        path: 'confirmNewPassword',
        message: 'Passwords do not match',
      })
    })
  })

  describe('Form Submission', () => {
    const validFormData = {
      otp: '123456',
      newPassword: 'password123',
      confirmNewPassword: 'password123',
    }

    beforeEach(async () => {
      const otpInput = component.find('input[name="otp"]')
      await otpInput.setValue(validFormData.otp)

      const newPasswordInput = component.find('input[name="newPassword"]')
      await newPasswordInput.setValue(validFormData.newPassword)

      const confirmNewPasswordInput = component.find(
        'input[name="confirmNewPassword"]'
      )
      await confirmNewPasswordInput.setValue(validFormData.confirmNewPassword)
    })

    it('calls reset password API with correct data', async () => {
      const form = component.find('form')
      await form.trigger('submit')

      expect(mockResetPassword).toHaveBeenCalledWith({
        email: mockAuthStore.forgotPassEmail,
        new_password: validFormData.newPassword,
        otp: validFormData.otp,
      })
    })

    it.skip('shows success state on completion', async () => {
      mockResetPassword.mockResolvedValueOnce({})

      const form = component.find('form')
      await form.trigger('submit')

      // @ts-expect-error vm does not detect ts types
      expect(component.vm.isResetComplete._value).toBe(true)
    })

    it.skip('shows error toast on invalid OTP', async () => {
      const error = new Error()
      Object.defineProperty(error, 'status', { value: 400 })
      mockResetPassword.mockRejectedValueOnce(error)

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockToast.add).toHaveBeenCalledWith({
        id: ToastId.RESET_PASSWORD_FAILED,
        color: 'red',
        title: 'Failed to reset password',
        description: 'Invalid OTP. Please try again.',
        icon: 'i-heroicons-exclamation-circle',
      })
    })
  })

  describe('Route Guard', () => {
    it('redirects to forgot password if email not set', async () => {
      mockAuthStore.forgotPassEmail = null
      await mountSuspended(ResetPasswordPage)
      expect(mockRouter.replace).toHaveBeenCalledWith('/auth/forgot-password')
    })
  })
})
