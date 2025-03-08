import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import SignUpPage from './signup.vue'

const { mockNavigateTo, mockSignUp, mockToast, mockAuthStore } = vi.hoisted(
  () => ({
    mockNavigateTo: vi.fn(),
    mockSignUp: vi.fn(),
    mockToast: {
      add: vi.fn(),
      clear: vi.fn(),
      remove: vi.fn(),
      update: vi.fn(),
    },
    mockAuthStore: {
      setSignUpEmail: vi.fn(),
      setForgotPassEmail: vi.fn(),
    },
  })
)

mockNuxtImport('navigateTo', () => mockNavigateTo)
mockNuxtImport('signUp', () => mockSignUp)
mockNuxtImport('useToast', () => () => mockToast)
mockNuxtImport('useAuthStore', () => () => mockAuthStore)

describe('SignUp Page', () => {
  let component: Awaited<ReturnType<typeof mountSuspended<typeof SignUpPage>>>

  beforeEach(async () => {
    component = await mountSuspended(SignUpPage)
    vi.clearAllMocks()
  })

  describe('Page Layout', () => {
    it('renders the signup form', () => {
      expect(component.find('form').exists()).toBe(true)
      expect(component.findAll('input[type="email"]')).toHaveLength(2)
      expect(component.findAll('input[type="password"]')).toHaveLength(2)
      expect(component.find('button[type="submit"]').exists()).toBe(true)
    })

    it('displays login link', () => {
      const link = component.find('a[href="/auth/login"]')
      expect(link.exists()).toBe(true)
      expect(link.text()).toContain('Login')
    })
  })

  describe('Form Validation', () => {
    it('shows error when emails do not match', async () => {
      const emailInput = component.find('input[name="email"]')
      await emailInput.setValue('test@example.com')

      const confirmEmailInput = component.find('input[name="confirmEmail"]')
      await confirmEmailInput.setValue('other@example.com')

      // @ts-expect-error vm does not detect ts types
      const errors = component.vm.validate(component.vm.signupFormData._value)
      expect(errors).toContainEqual({
        path: 'confirmEmail',
        message: 'Emails do not match',
      })
    })

    it('shows error when passwords do not match', async () => {
      const passwordInput = component.find('input[name="password"]')
      await passwordInput.setValue('password123')

      const confirmPasswordInput = component.find(
        'input[name="confirmPassword"]'
      )
      await confirmPasswordInput.setValue('password456')

      // @ts-expect-error vm does not detect ts types
      const errors = component.vm.validate(component.vm.signupFormData._value)
      expect(errors).toContainEqual({
        path: 'confirmPassword',
        message: 'Passwords do not match',
      })
    })
  })

  describe('Form Submission', () => {
    const validFormData = {
      email: 'test@example.com',
      confirmEmail: 'test@example.com',
      password: 'password123',
      confirmPassword: 'password123',
    }

    beforeEach(async () => {
      const emailInput = component.find('input[name="email"]')
      await emailInput.setValue(validFormData.email)

      const confirmEmailInput = component.find('input[name="confirmEmail"]')
      await confirmEmailInput.setValue(validFormData.confirmEmail)

      const passwordInput = component.find('input[name="password"]')
      await passwordInput.setValue(validFormData.password)

      const confirmPasswordInput = component.find(
        'input[name="confirmPassword"]'
      )
      await confirmPasswordInput.setValue(validFormData.confirmPassword)
    })

    it('calls signup API with correct data', async () => {
      const form = component.find('form')
      await form.trigger('submit')

      expect(mockSignUp).toHaveBeenCalledWith({
        email: validFormData.email,
        password: validFormData.password,
      })
    })

    it.skip('stores email and navigates to verification on success', async () => {
      mockSignUp.mockResolvedValueOnce({})

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockAuthStore.setSignUpEmail).toHaveBeenCalledWith(
        validFormData.email
      )
      expect(mockNavigateTo).toHaveBeenCalledWith('/auth/verification')
    })

    it.skip('shows error toast when email already exists', async () => {
      const error = new Error()
      Object.defineProperty(error, 'status', { value: 409 })
      mockSignUp.mockRejectedValueOnce(error)

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockToast.add).toHaveBeenCalledWith({
        id: ToastId.SIGNUP_FAILED,
        color: 'red',
        title: 'Failed to create account',
        description: 'This email is already registered.',
        icon: 'i-heroicons-exclamation-circle',
      })
    })
  })
})
