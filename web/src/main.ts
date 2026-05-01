import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useUserStore } from './stores/user'
import { setupAuthCallbacks } from './api/axios'

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