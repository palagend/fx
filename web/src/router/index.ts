import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'
import { defineAsyncComponent } from 'vue'

// 首屏关键组件同步加载
import Home from '../components/Home.vue'

// 其他组件懒加载，减少首屏 JS 大小
const ExchangeRate = defineAsyncComponent(() => import('../components/ExchangeRate.vue'))
const Calculator = defineAsyncComponent(() => import('../components/Calculator.vue'))
const Portfolio = defineAsyncComponent(() => import('../components/Portfolio.vue'))
const QRCodeGenerator = defineAsyncComponent(() => import('../components/QRCodeGenerator.vue'))
const PasswordGenerator = defineAsyncComponent(() => import('../components/PasswordGenerator.vue'))
const PasswordManager = defineAsyncComponent(() => import('../components/PasswordManager.vue'))
const MobileTools = defineAsyncComponent(() => import('../components/MobileTools.vue'))
const MobileProfile = defineAsyncComponent(() => import('../components/MobileProfile.vue'))

const routes: RouteRecordRaw[] = [
  { 
    path: '/', 
    name: 'Home',
    component: Home,
    beforeEnter: (to, from, next) => {
      if (typeof window !== 'undefined' && window.innerWidth < 768) {
        next('/portfolio')
      } else {
        next()
      }
    }
  },
  { path: '/tools', component: MobileTools },
  { path: '/exchange-rate', component: ExchangeRate },
  { path: '/calculator', component: Calculator },
  { path: '/portfolio', component: Portfolio },
  { path: '/qrcode-generator', component: QRCodeGenerator },
  { path: '/password-generator', component: PasswordGenerator },
  { path: '/password-manager', component: PasswordManager },
  { path: '/profile', component: MobileProfile }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
