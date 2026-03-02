// Этот файл Nuxt импортирует автоматически!
export const useAPI = <T>(url: string, options: any = {}) => {
  const config = useRuntimeConfig()

  // Магия Docker:
  // Если мы на сервере -> берем внутренний адрес (http://atm_backend:8080)
  // Если в браузере -> берем внешний адрес (http://localhost:8080)
  const baseURL = import.meta.server 
    ? config.apiBase 
    : config.public.apiBase

  return useFetch<T>(url, {
    baseURL,
    key: url, // Уникальный ключ для кэширования
    ...options
  })
}