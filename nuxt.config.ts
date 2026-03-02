// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  runtimeConfig: {
    public: {
      // Это на будущее, если захочешь менять URL
      apiBase: 'http://localhost:8080'
    }
  },

  nitro: {
    devProxy: {
      '/api': {
        target: 'http://localhost:8080/api',
        changeOrigin: true,
        prependPath: true
      }
    }
  },

  devtools: { enabled: false },

  sourcemap: {
    server: false,
    client: false
  },

  typescript: {
    strict: false,
    typeCheck: false
  },

  compatibilityDate: '2025-01-18'
})