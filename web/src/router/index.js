import { createRouter, createWebHistory } from 'vue-router'
import Home from '../components/Home.vue'
import ExchangeRate from '../components/ExchangeRate.vue'
import Calculator from '../components/Calculator.vue'
import CryptoPortfolio from '../components/CryptoPortfolio.vue'
import QRCodeGenerator from '../components/QRCodeGenerator.vue'
import PasswordGenerator from '../components/PasswordGenerator.vue'
import PasswordManager from '../components/PasswordManager.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/exchange-rate', component: ExchangeRate },
  { path: '/calculator', component: Calculator },
  { path: '/crypto-portfolio', component: CryptoPortfolio },
  { path: '/qrcode-generator', component: QRCodeGenerator },
  { path: '/password-generator', component: PasswordGenerator },
  { path: '/password-manager', component: PasswordManager }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
