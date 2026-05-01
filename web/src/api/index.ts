import { config } from '../config'
import { backendPortfolioApi } from './portfolio'
import { localPortfolioApi } from './localStorage/portfolio'
import { localUserApi } from './localStorage/user'
import * as localPasswordApi from './localStorage/password'

export const portfolioApi = config.isBackend ? backendPortfolioApi : localPortfolioApi

export const userApi = config.isBackend ? null : localUserApi

export const passwordApi = localPasswordApi

export { apiClient } from './axios'

export { config }