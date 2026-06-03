import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useUserStore } from './stores/user'
import { setupAuthCallbacks } from './api/axios'

// GitHub Pages 404 重定向处理
const redirect = sessionStorage.getItem('redirect')
if (redirect) {
  sessionStorage.removeItem('redirect')
  const basePath = '/' + location.pathname.split('/')[1]
  const newPath = redirect.replace(basePath, '') || '/'
  history.replaceState(null, '', newPath)
}

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const userStore = useUserStore()
userStore.init()

setupAuthCallbacks(
  () => userStore.accessToken,
  () => userStore.refreshToken,
  (tokens) => userStore.setTokens(tokens),
  () => userStore.clearTokens(),
  () => userStore.openLoginModal()
)

app.mount('#app')