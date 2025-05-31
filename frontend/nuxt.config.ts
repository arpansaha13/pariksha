// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@nuxt/ui',
    'reka-ui/nuxt',
    '@vueuse/nuxt',
    '@pinia/nuxt',
    '@nuxt/image',
    '@nuxt/eslint',
    '@nuxt/test-utils/module',
    'pinia-plugin-persistedstate/nuxt',
    'nuxt-monaco-editor',
    'nuxt-lodash',
    // '@nuxtjs/html-validator',
  ],

  // https://github.com/nuxt/nuxt/issues/12003#issuecomment-1397230032
  vite: {
    server: {
      hmr: {
        path: 'hmr/',
      },
    },
  },

  runtimeConfig: {
    env: '',
    apiBaseUrl: '',
  },

  ui: {
    colorMode: false,
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

  css: ['~/assets/css/main.css'],

  image: {
    format: ['webp'],
    densities: [1, 2],
  },

  lodash: {
    prefix: '_',
    prefixSkip: false,
    upperAfterPrefix: false,
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
