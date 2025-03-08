import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { sleep } from '~/tests/utils'
import LoginPage from './login.vue'

const { mockNavigateTo, mockLogin, mockToast } = vi.hoisted(() => {
  return {
    mockLogin: vi.fn(),
    mockNavigateTo: vi.fn(),
    mockToast: {
      add: vi.fn(),
    },
  }
})

mockNuxtImport('login', () => mockLogin)
mockNuxtImport('navigateTo', () => mockNavigateTo)
mockNuxtImport('useToast', () => () => mockToast)

describe('Login Page', () => {
  let component: Awaited<ReturnType<typeof mountSuspended<typeof LoginPage>>>

  beforeEach(async () => {
    component = await mountSuspended(LoginPage)
    vi.clearAllMocks()
  })

  describe('Page Layout', () => {
    it('renders the login form', () => {
      expect(component.find('form').exists()).toBe(true)
      expect(component.find('input[type="email"]').exists()).toBe(true)
      expect(component.find('input[type="password"]').exists()).toBe(true)
      expect(component.find('button[type="submit"]').exists()).toBe(true)
    })

    it('displays correct navigation links', () => {
      const links = component.findAll('a')
      expect(
        links.some(link => link.attributes('href') === '/auth/signup')
      ).toBe(true)
      expect(
        links.some(link => link.attributes('href') === '/auth/forgot-password')
      ).toBe(true)
    })
  })

  describe('Form Functionality', () => {
    it('updates email value on input', async () => {
      const email = 'test@example.com'
      const emailInput = component.find('input[type="email"]')
      await emailInput.setValue(email)

      // @ts-expect-error vm does not detect ts types
      expect(component.vm.loginFormData._value.email).toBe(email)
    })

    it('updates password value on input', async () => {
      const password = 'password123'
      const passwordInput = component.find('input[type="password"]')
      await passwordInput.setValue(password)

      // @ts-expect-error vm does not detect ts types
      expect(component.vm.loginFormData._value.password).toBe(password)
    })

    it('requires email and password fields', () => {
      const emailInput = component.find('input[type="email"]')
      const passwordInput = component.find('input[type="password"]')

      expect(emailInput.attributes('required')).toBeDefined()
      expect(passwordInput.attributes('required')).toBeDefined()
    })
  })

  describe('Form Submission', () => {
    const validCredentials = {
      email: 'test@example.com',
      password: 'password123',
    }

    beforeEach(async () => {
      const emailInput = component.find('input[type="email"]')
      await emailInput.setValue(validCredentials.email)

      const passwordInput = component.find('input[type="password"]')
      await passwordInput.setValue(validCredentials.password)
    })

    it.skip('shows loading state during submission', async () => {
      mockLogin.mockImplementationOnce(() => sleep(100))

      const form = component.find('form')
      form.trigger('submit')

      // @ts-expect-error vm does not detect ts types
      expect(component.vm.loading._value).toBe(true)
    })

    it('calls login API with correct credentials', async () => {
      const form = component.find('form')
      await form.trigger('submit')

      expect(mockLogin).toHaveBeenCalledWith(validCredentials)
    })

    it.skip('navigates to home on successful login', async () => {
      mockLogin.mockResolvedValueOnce({})

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockNavigateTo).toHaveBeenCalledWith('/')
    })

    it.skip('shows error toast on login failure', async () => {
      mockLogin.mockRejectedValueOnce(new Error('Login failed'))

      const form = component.find('form')
      await form.trigger('submit')

      expect(mockToast.add).toHaveBeenCalledWith({
        id: ToastId.LOGIN_FAILED,
        color: 'red',
        title: 'Failed to login',
        description: 'Invalid email or password.',
        icon: 'i-heroicons-exclamation-circle',
      })
    })

    it.skip('resets loading state after submission', async () => {
      const form = component.find('form')
      await form.trigger('submit')

      // @ts-expect-error vm does not detect ts types
      expect(component.vm.loading._value).toBe(false)
    })
  })
})
