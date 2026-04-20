// LocalStorage 版本的用户 API（纯前端模式）

const STORAGE_KEYS = {
  USER: 'local_user',
  IS_LOGGED_IN: 'local_is_logged_in'
}

// 模拟用户数据
let mockUser = null

// 模拟 API 响应格式
function mockResponse(data, delay = 300) {
  return new Promise(resolve => {
    setTimeout(() => resolve({ data }), delay)
  })
}

function mockError(message, delay = 300) {
  return new Promise((_, reject) => {
    setTimeout(() => {
      reject({
        response: { data: { error: message } }
      })
    }, delay)
  })
}

export const localUserApi = {
  // 注册
  async register(username, email, password) {
    const existingUser = localStorage.getItem(STORAGE_KEYS.USER)
    if (existingUser) {
      const user = JSON.parse(existingUser)
      if (user.username === username) {
        return mockError('用户名已存在')
      }
    }

    mockUser = {
      id: Date.now().toString(),
      username,
      email,
      created_at: new Date().toISOString()
    }

    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(mockUser))
    localStorage.setItem(STORAGE_KEYS.IS_LOGGED_IN, 'true')

    return mockResponse({
      user: mockUser,
      tokens: {
        access_token: 'local_mock_token',
        refresh_token: 'local_mock_refresh_token'
      }
    })
  },

  // 登录
  async login(username, password) {
    const storedUser = localStorage.getItem(STORAGE_KEYS.USER)

    if (!storedUser) {
      return mockError('用户不存在，请先注册')
    }

    const user = JSON.parse(storedUser)
    if (user.username !== username) {
      return mockError('用户名或密码错误')
    }

    mockUser = user
    localStorage.setItem(STORAGE_KEYS.IS_LOGGED_IN, 'true')

    return mockResponse({
      user: mockUser,
      tokens: {
        access_token: 'local_mock_token',
        refresh_token: 'local_mock_refresh_token'
      }
    })
  },

  // 获取当前用户
  async getCurrentUser() {
    const isLoggedIn = localStorage.getItem(STORAGE_KEYS.IS_LOGGED_IN) === 'true'
    if (!isLoggedIn) {
      return mockError('未登录')
    }

    const storedUser = localStorage.getItem(STORAGE_KEYS.USER)
    if (storedUser) {
      mockUser = JSON.parse(storedUser)
      return mockResponse({ user: mockUser })
    }

    return mockError('用户不存在')
  },

  // 登出
  async logout() {
    localStorage.setItem(STORAGE_KEYS.IS_LOGGED_IN, 'false')
    mockUser = null
    return mockResponse({ success: true })
  },

  // 登出所有设备
  async logoutAll() {
    localStorage.setItem(STORAGE_KEYS.IS_LOGGED_IN, 'false')
    mockUser = null
    return mockResponse({ success: true })
  },

  // 修改密码
  async changePassword(oldPassword, newPassword) {
    return mockResponse({ success: true })
  },

  // 刷新 token
  async refreshToken(refreshToken) {
    return mockResponse({
      tokens: {
        access_token: 'local_mock_token',
        refresh_token: 'local_mock_refresh_token'
      }
    })
  },

  // 检查登录状态
  isLoggedIn() {
    return localStorage.getItem(STORAGE_KEYS.IS_LOGGED_IN) === 'true'
  },

  // 获取本地用户
  getUser() {
    const storedUser = localStorage.getItem(STORAGE_KEYS.USER)
    return storedUser ? JSON.parse(storedUser) : null
  }
}
