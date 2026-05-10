import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'

// 首屏关键组件同步加载
import Home from '../components/Home.vue'

// 其他组件懒加载，使用具名 chunk 合并相关模块
const ExchangeRate = () => import(/* webpackChunkName: "feature-tools" */ '../components/ExchangeRate.vue')
const Calculator = () => import(/* webpackChunkName: "feature-tools" */ '../components/Calculator.vue')
const QRCodeGenerator = () => import(/* webpackChunkName: "feature-tools" */ '../components/QRCodeGenerator.vue')

const Portfolio = () => import(/* webpackChunkName: "feature-portfolio" */ '../components/Portfolio.vue')

const PasswordGenerator = () => import(/* webpackChunkName: "feature-password" */ '../components/PasswordGenerator.vue')
const PasswordManager = () => import(/* webpackChunkName: "feature-password" */ '../components/PasswordManager.vue')

const MobileTools = () => import(/* webpackChunkName: "feature-mobile" */ '../components/MobileTools.vue')
const MobileProfile = () => import(/* webpackChunkName: "feature-mobile" */ '../components/MobileProfile.vue')

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
