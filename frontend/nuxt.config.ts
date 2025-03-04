// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@vueuse/nuxt',
    '@pinia/nuxt',
    '@nuxt/ui',
    '@nuxt/image',
    '@nuxt/eslint',
    '@nuxt/fonts',
    '@nuxt/test-utils/module',
    'nuxt-headlessui',
    'pinia-plugin-persistedstate/nuxt',
    // '@nuxtjs/html-validator',
  ],

  runtimeConfig: {
    apiBaseUrl: '',
    public: {
      apiBaseUrl: '',
    },
  },

  srcDir: 'src/',

  app: {
    head: {
      htmlAttrs: {
        lang: 'en',
      },
      link: [{ rel: 'icon', type: 'image/png', href: '/favicon.ico' }],
    },
  },

  css: ['assets/main.css'],

  ui: {
    global: true,
  },

  image: {
    format: ['webp'],
    densities: [1, 2],
  },

  headlessui: {
    prefix: 'Headless',
  },

  colorMode: {
    preference: 'light',
    classPrefix: '',
    classSuffix: '',
  },

  // htmlValidator: {
  //   usePrettier: false,
  //   logLevel: 'verbose',
  //   failOnError: false,
  //   options: {
  //     extends: [
  //       'html-validate:document',
  //       'html-validate:recommended',
  //       'html-validate:standard',
  //     ],
  //     rules: {
  //       'svg-focusable': 'off',
  //       'no-unknown-elements': 'error',
  //       // Conflicts or not needed as we use prettier formatting
  //       'void-style': 'off',
  //       'no-trailing-whitespace': 'off',
  //       // Conflict with Nuxt defaults
  //       'require-sri': 'off',
  //       'attribute-boolean-style': 'off',
  //       'doctype-style': 'off',
  //       // Unreasonable rule
  //       'no-inline-style': 'off',
  //     },
  //   },
  // },

  compatibilityDate: '2025-03-01',
})
