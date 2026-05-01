import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import Home from '../components/Home.vue'
import ExchangeRate from '../components/ExchangeRate.vue'
import Calculator from '../components/Calculator.vue'
import Portfolio from '../components/Portfolio.vue'
import QRCodeGenerator from '../components/QRCodeGenerator.vue'
import PasswordGenerator from '../components/PasswordGenerator.vue'
import PasswordManager from '../components/PasswordManager.vue'
import MobileTools from '../components/MobileTools.vue'
import MobileProfile from '../components/MobileProfile.vue'

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
  history: createWebHistory(),
  routes
})

export default router