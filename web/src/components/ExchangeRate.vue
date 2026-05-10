<template>
  <div class="exchange-rate-container">
    <!-- 头部标题 -->
    <header class="page-header">
      <h1 class="page-title">
        <Icon icon="mdi:swap-horizontal" />
        <span>汇率换算</span>
      </h1>
    </header>

    <!-- 主换算卡片 -->
    <div class="main-card">
      <!-- 货币选择行 -->
      <div class="currency-selector">
        <div class="currency-box" @click="showFromSelector = true">
          <span class="currency-flag">{{ getCurrencyFlag(from) }}</span>
          <div class="currency-info">
            <span class="currency-code">{{ from }}</span>
            <span class="currency-name">{{ getCurrencyName(from) }}</span>
          </div>
          <Icon icon="mdi:chevron-right" class="arrow-icon" />
        </div>

        <button class="swap-btn" @click="swapCurrency" :class="{ swapping: isSwapping }">
          <Icon icon="mdi:swap-horizontal" />
        </button>

        <div class="currency-box" @click="showToSelector = true">
          <span class="currency-flag">{{ getCurrencyFlag(to) }}</span>
          <div class="currency-info">
            <span class="currency-code">{{ to }}</span>
            <span class="currency-name">{{ getCurrencyName(to) }}</span>
          </div>
          <Icon icon="mdi:chevron-right" class="arrow-icon" />
        </div>
      </div>

      <!-- 汇率信息条 -->
      <div class="rate-bar" v-if="rateFixed">
        <span class="rate-label">实时汇率</span>
        <span class="rate-value">1 {{ from }} = {{ rateFixed }} {{ to }}</span>
      </div>

      <!-- 金额输入区 -->
      <div class="amount-section">
        <div class="amount-input-box">
          <span class="amount-symbol">{{ getCurrencySymbol(from) }}</span>
          <input
            type="number"
            v-model.number="amount"
            placeholder="0.00"
            class="amount-input"
            @input="result = ''; rateFixed = ''"
          />
          <button v-if="amount" class="clear-btn" @click="clearAmount">
            <Icon icon="mdi:close-circle" />
          </button>
        </div>
        <div class="amount-hint">请输入兑换金额</div>
      </div>

      <!-- 快捷金额 -->
      <div class="quick-amounts">
        <button
          v-for="amt in quickAmounts"
          :key="amt"
          class="quick-amount-btn"
          :class="{ active: amount === amt }"
          @click="setAmount(amt)"
        >
          {{ formatCompactNumber(amt) }}
        </button>
      </div>

      <!-- 换算按钮 -->
      <button class="convert-btn" @click="getRate" :disabled="loading || !amount">
        <Icon :icon="loading ? 'mdi:loading' : 'mdi:calculator-variant'" />
        <span>{{ loading ? '计算中...' : '立即换算' }}</span>
      </button>
    </div>

    <!-- 结果展示卡片 -->
    <div ref="resultCardRef" class="result-card" v-if="result" @click="copyResult">
      <div class="result-header">
        <span class="result-label">换算结果</span>
        <button class="copy-btn" :class="{ copied: copySuccess }" @click.stop="copyResult">
          <Icon :icon="copySuccess ? 'mdi:check' : 'mdi:content-copy'" />
        </button>
      </div>
      <div class="result-body">
        <div class="result-from">
          <span class="result-amount">{{ formatNumber(amount) }}</span>
          <span class="result-currency">{{ from }}</span>
        </div>
        <div class="result-equals">
          <Icon icon="mdi:arrow-down" />
        </div>
        <div class="result-to">
          <span class="result-amount highlight">{{ formatNumber(result) }}</span>
          <span class="result-currency">{{ to }}</span>
        </div>
      </div>
      <div class="result-footer" v-if="updateTime">
        <Icon icon="mdi:clock-outline" />
        <span>更新时间：{{ updateTime }}</span>
      </div>
    </div>

    <!-- 损耗计算卡片 -->
    <div class="loss-card" v-if="result">
      <div class="loss-header" @click="toggleLossSection">
        <div class="loss-title">
          <Icon icon="mdi:chart-line" />
          <span>兑换损耗分析</span>
        </div>
        <Icon :icon="showLossSection ? 'mdi:chevron-up' : 'mdi:chevron-down'" class="toggle-icon" />
      </div>

      <transition name="slide">
        <div v-show="showLossSection" class="loss-body">
          <div class="loss-input-box">
            <label>实际到账金额 ({{ to }})</label>
            <div class="loss-input-wrapper">
              <span class="loss-symbol">{{ getCurrencySymbol(to) }}</span>
              <input
                type="number"
                v-model.number="actualAmount"
                placeholder="0.00"
                @keyup.enter="calcLoss"
              />
            </div>
          </div>

          <button class="calc-loss-btn" @click="calcLoss" :disabled="!actualAmount">
            <Icon icon="mdi:calculator" />
            <span>计算损耗</span>
          </button>

          <div class="loss-result" v-if="lossData.show">
            <div class="loss-grid">
              <div class="loss-item">
                <span class="loss-item-label">本次损耗</span>
                <span class="loss-item-value">{{ formatNumber(lossData.currentLoss) }} {{ to }}</span>
              </div>
              <div class="loss-item">
                <span class="loss-item-label">每万元损耗</span>
                <span class="loss-item-value">{{ formatNumber(lossData.per10000Loss) }} {{ to }}</span>
              </div>
              <div class="loss-item highlight">
                <span class="loss-item-label">损耗率</span>
                <span class="loss-item-value">{{ lossData.lossRate.toFixed(2) }}%</span>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </div>

    <!-- FAB 历史记录按钮 -->
    <button
      v-if="history.length > 0"
      class="fab-history-btn"
      @click="showHistoryModal = true"
      title="最近查询"
    >
      <Icon icon="mdi:history" />
      <span class="fab-badge" v-if="history.length > 0">{{ Math.min(history.length, 6) }}</span>
    </button>

    <!-- 历史记录弹窗 -->
    <div v-if="showHistoryModal" class="history-modal-overlay" @click="showHistoryModal = false">
      <div class="history-modal" @click.stop>
        <div class="history-modal-header">
          <h3>
            <Icon icon="mdi:history" />
            <span>最近查询</span>
          </h3>
          <button class="close-btn" @click="showHistoryModal = false">
            <Icon icon="mdi:close" />
          </button>
        </div>
        <div class="history-modal-body">
          <div class="history-list">
            <div
              v-for="(item, index) in history.slice(0, 6)"
              :key="index"
              class="history-item"
              @click="loadHistory(item)"
            >
              <div class="history-currency-pair">
                <span class="history-flag">{{ getCurrencyFlag(item.from) }}</span>
                <span class="history-code">{{ item.from }}</span>
                <Icon icon="mdi:arrow-right" class="history-arrow" />
                <span class="history-flag">{{ getCurrencyFlag(item.to) }}</span>
                <span class="history-code">{{ item.to }}</span>
              </div>
              <div class="history-details">
                <span class="history-amount">{{ formatNumber(item.amount) }} → {{ formatNumber(item.result) }}</span>
                <span class="history-rate">{{ item.rate }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="history-modal-footer">
          <button class="clear-history-btn" @click="clearHistory">
            <Icon icon="mdi:trash-can-outline" />
            <span>清空记录</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 货币选择弹窗 - 源货币 -->
    <div v-if="showFromSelector" class="currency-modal-overlay" @click="showFromSelector = false">
      <div class="currency-modal" @click.stop>
        <div class="currency-modal-header">
          <h3>选择源货币</h3>
          <button class="close-btn" @click="showFromSelector = false">
            <Icon icon="mdi:close" />
          </button>
        </div>
        <div class="currency-search">
          <Icon icon="mdi:magnify" />
          <input
            type="text"
            v-model="currencySearch"
            placeholder="搜索货币名称或代码"
          />
        </div>
        <div class="currency-list">
          <div class="currency-group">
            <div class="group-title">常用货币</div>
            <div
              v-for="currency in filteredCommonCurrencies"
              :key="currency.code"
              class="currency-option"
              :class="{ active: from === currency.code }"
              @click="selectFromCurrency(currency.code)"
            >
              <span class="currency-flag">{{ currency.flag }}</span>
              <div class="currency-info">
                <span class="currency-code">{{ currency.code }}</span>
                <span class="currency-name">{{ currency.name }}</span>
              </div>
              <Icon v-if="from === currency.code" icon="mdi:check" class="check-icon" />
            </div>
          </div>
          <div class="currency-group" v-if="filteredOtherCurrencies.length">
            <div class="group-title">其他货币</div>
            <div
              v-for="currency in filteredOtherCurrencies"
              :key="currency.code"
              class="currency-option"
              :class="{ active: from === currency.code }"
              @click="selectFromCurrency(currency.code)"
            >
              <span class="currency-flag">{{ currency.flag }}</span>
              <div class="currency-info">
                <span class="currency-code">{{ currency.code }}</span>
                <span class="currency-name">{{ currency.name }}</span>
              </div>
              <Icon v-if="from === currency.code" icon="mdi:check" class="check-icon" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 货币选择弹窗 - 目标货币 -->
    <div v-if="showToSelector" class="currency-modal-overlay" @click="showToSelector = false">
      <div class="currency-modal" @click.stop>
        <div class="currency-modal-header">
          <h3>选择目标货币</h3>
          <button class="close-btn" @click="showToSelector = false">
            <Icon icon="mdi:close" />
          </button>
        </div>
        <div class="currency-search">
          <Icon icon="mdi:magnify" />
          <input
            type="text"
            v-model="currencySearch"
            placeholder="搜索货币名称或代码"
          />
        </div>
        <div class="currency-list">
          <div class="currency-group">
            <div class="group-title">常用货币</div>
            <div
              v-for="currency in filteredCommonCurrencies"
              :key="currency.code"
              class="currency-option"
              :class="{ active: to === currency.code }"
              @click="selectToCurrency(currency.code)"
            >
              <span class="currency-flag">{{ currency.flag }}</span>
              <div class="currency-info">
                <span class="currency-code">{{ currency.code }}</span>
                <span class="currency-name">{{ currency.name }}</span>
              </div>
              <Icon v-if="to === currency.code" icon="mdi:check" class="check-icon" />
            </div>
          </div>
          <div class="currency-group" v-if="filteredOtherCurrencies.length">
            <div class="group-title">其他货币</div>
            <div
              v-for="currency in filteredOtherCurrencies"
              :key="currency.code"
              class="currency-option"
              :class="{ active: to === currency.code }"
              @click="selectToCurrency(currency.code)"
            >
              <span class="currency-flag">{{ currency.flag }}</span>
              <div class="currency-info">
                <span class="currency-code">{{ currency.code }}</span>
                <span class="currency-name">{{ currency.name }}</span>
              </div>
              <Icon v-if="to === currency.code" icon="mdi:check" class="check-icon" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Toast 提示 -->
    <div class="toast" :class="{ show: toast.show }">
      <Icon :icon="toast.icon" />
      <span>{{ toast.text }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import axios from 'axios'
import { Icon } from '@iconify/vue'

// 响应式数据
const amount = ref('')
const from = ref('USD')
const to = ref('CNY')
const result = ref('')
const rateFixed = ref('')
const updateTime = ref('')
const actualAmount = ref('')
const loading = ref(false)
const showLossSection = ref(false)
const copySuccess = ref(false)
const showFromSelector = ref(false)
const showToSelector = ref(false)
const showHistoryModal = ref(false)
const currencySearch = ref('')
const isSwapping = ref(false)
const resultCardRef = ref(null)

const lossData = ref({
  show: false,
  currentLoss: 0,
  per10000Loss: 0,
  lossRate: 0
})

const toast = ref({
  show: false,
  text: '',
  icon: 'mdi:check'
})

const history = ref([])

// 货币数据
const commonCurrencies = ref([
  { code: 'USD', name: '美元', flag: '🇺🇸', symbol: '$' },
  { code: 'CNY', name: '人民币', flag: '🇨🇳', symbol: '¥' },
  { code: 'HKD', name: '港币', flag: '🇭🇰', symbol: '$' },
  { code: 'EUR', name: '欧元', flag: '🇪🇺', symbol: '€' },
  { code: 'GBP', name: '英镑', flag: '🇬🇧', symbol: '£' },
  { code: 'JPY', name: '日元', flag: '🇯🇵', symbol: '¥' }
])

const otherCurrencies = ref([
  { code: 'AUD', name: '澳元', flag: '🇦🇺', symbol: '$' },
  { code: 'CAD', name: '加元', flag: '🇨🇦', symbol: '$' },
  { code: 'CHF', name: '瑞士法郎', flag: '🇨🇭', symbol: 'Fr' },
  { code: 'SGD', name: '新加坡元', flag: '🇸🇬', symbol: '$' },
  { code: 'KRW', name: '韩元', flag: '🇰🇷', symbol: '₩' },
  { code: 'THB', name: '泰铢', flag: '🇹🇭', symbol: '฿' },
  { code: 'RUB', name: '卢布', flag: '🇷🇺', symbol: '₽' },
  { code: 'INR', name: '印度卢比', flag: '🇮🇳', symbol: '₹' },
  { code: 'MYR', name: '马来西亚林吉特', flag: '🇲🇾', symbol: 'RM' },
  { code: 'VND', name: '越南盾', flag: '🇻🇳', symbol: '₫' },
  { code: 'NZD', name: '新西兰元', flag: '🇳🇿', symbol: '$' },
  { code: 'SEK', name: '瑞典克朗', flag: '🇸🇪', symbol: 'kr' },
  { code: 'NOK', name: '挪威克朗', flag: '🇳🇴', symbol: 'kr' },
  { code: 'DKK', name: '丹麦克朗', flag: '🇩🇰', symbol: 'kr' },
  { code: 'PLN', name: '波兰兹罗提', flag: '🇵🇱', symbol: 'zł' },
  { code: 'MXN', name: '墨西哥比索', flag: '🇲🇽', symbol: '$' },
  { code: 'BRL', name: '巴西雷亚尔', flag: '🇧🇷', symbol: 'R$' },
  { code: 'ZAR', name: '南非兰特', flag: '🇿🇦', symbol: 'R' },
  { code: 'AED', name: '阿联酋迪拉姆', flag: '🇦🇪', symbol: 'د.إ' },
  { code: 'SAR', name: '沙特里亚尔', flag: '🇸🇦', symbol: '﷼' }
])

// 快捷金额
const quickAmounts = [100, 500, 1000, 5000, 10000, 50000]

// 计算属性
const allCurrencies = computed(() => [...commonCurrencies.value, ...otherCurrencies.value])

const filteredCommonCurrencies = computed(() => {
  if (!currencySearch.value) return commonCurrencies.value
  const search = currencySearch.value.toLowerCase()
  return commonCurrencies.value.filter(
    c => c.code.toLowerCase().includes(search) || c.name.includes(search)
  )
})

const filteredOtherCurrencies = computed(() => {
  if (!currencySearch.value) return otherCurrencies.value
  const search = currencySearch.value.toLowerCase()
  return otherCurrencies.value.filter(
    c => c.code.toLowerCase().includes(search) || c.name.includes(search)
  )
})

// 方法
const getCurrencyFlag = (code) => {
  const currency = allCurrencies.value.find(c => c.code === code)
  return currency?.flag || '🌍'
}

const getCurrencyName = (code) => {
  const currency = allCurrencies.value.find(c => c.code === code)
  return currency?.name || code
}

const getCurrencySymbol = (code) => {
  const currency = allCurrencies.value.find(c => c.code === code)
  return currency?.symbol || code
}

const formatNumber = (num) => {
  if (!num && num !== 0) return '0'
  return Number(num).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
}

const formatCompactNumber = (num) => {
  if (num >= 10000) {
    return (num / 10000).toFixed(0) + '万'
  }
  return num.toLocaleString('zh-CN')
}

const setAmount = async (amt) => {
  amount.value = amt
  showToast(`已选择 ${formatCompactNumber(amt)}`, 'mdi:check')
  // 快捷金额自动换算
  await getRate()
}

const clearAmount = () => {
  amount.value = ''
  result.value = ''
}

const swapCurrency = () => {
  isSwapping.value = true
  const temp = from.value
  from.value = to.value
  to.value = temp
  result.value = ''
  rateFixed.value = ''
  setTimeout(() => {
    isSwapping.value = false
  }, 300)
}

const selectFromCurrency = (code) => {
  from.value = code
  showFromSelector.value = false
  currencySearch.value = ''
  result.value = ''
  rateFixed.value = ''
}

const selectToCurrency = (code) => {
  to.value = code
  showToSelector.value = false
  currencySearch.value = ''
  result.value = ''
  rateFixed.value = ''
}

const toggleLossSection = () => {
  showLossSection.value = !showLossSection.value
}

const showToast = (text, icon = 'mdi:check') => {
  toast.value = { show: true, text, icon }
  setTimeout(() => {
    toast.value.show = false
  }, 2000)
}

const getRate = async () => {
  if (!amount.value || amount.value <= 0) {
    showToast('请输入有效金额', 'mdi:alert-circle')
    return
  }

  loading.value = true

  try {
    const primaryApiUrl = 'https://api.exchangerate.host/live'
    const backupApiUrl = 'https://open.er-api.com/v6/latest/' + from.value
    const apiKey = '750a5c496977bdfc9770fa43bd914d07'

    let rate = null
    let apiTime = null

    try {
      const targetCurrencies = to.value
      const primaryUrl = `${primaryApiUrl}?access_key=${apiKey}&source=${from.value}&currencies=${targetCurrencies}`
      const response = await axios.get(primaryUrl)
      const data = response.data

      if (data.success) {
        const rateKey = from.value + to.value
        rate = data.quotes?.[rateKey]
        apiTime = data.timestamp
      }
    } catch (primaryError) {
      console.log('Primary API failed:', primaryError.message)
    }

    if (!rate) {
      try {
        const backupResponse = await axios.get(backupApiUrl)
        const backupData = backupResponse.data

        if (backupData.rates && backupData.rates[to.value]) {
          rate = backupData.rates[to.value]
          apiTime = backupData.time_last_update_unix
            ? backupData.time_last_update_unix
            : null
        }
      } catch (backupError) {
        console.error('Backup API failed:', backupError.message)
      }
    }

    if (!rate) {
      showToast('获取汇率失败，请重试', 'mdi:close-circle')
      loading.value = false
      return
    }

    rateFixed.value = rate.toFixed(6)
    result.value = (amount.value * rate).toFixed(2)

    const time = apiTime ? new Date(apiTime * 1000) : new Date()
    updateTime.value = time.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })

    saveToHistory()
    showToast('换算成功', 'mdi:check')

    // 延迟滚动到结果卡片，等待DOM更新
    setTimeout(() => {
      if (resultCardRef.value) {
        resultCardRef.value.scrollIntoView({
          behavior: 'smooth',
          block: 'center'
        })
      }
    }, 100)
  } catch (err) {
    showToast('网络错误，请检查连接', 'mdi:close-circle')
  } finally {
    loading.value = false
  }
}

const calcLoss = () => {
  if (!actualAmount.value || actualAmount.value <= 0) {
    showToast('请输入实际到账金额', 'mdi:alert-circle')
    return
  }

  const currentLoss = Number(result.value) - actualAmount.value
  if (currentLoss < 0) {
    showToast('实际到账不能高于换算结果', 'mdi:alert-circle')
    return
  }

  const lossRate = (currentLoss / Number(result.value)) * 100
  const per10000Loss = (currentLoss / Number(amount.value)) * 10000

  lossData.value = {
    show: true,
    currentLoss,
    per10000Loss,
    lossRate
  }

  showToast('损耗计算完成', 'mdi:chart-line')
}

const copyResult = () => {
  if (!result.value) return
  const text = `${amount.value} ${from.value} = ${result.value} ${to.value}`
  navigator.clipboard.writeText(text).then(() => {
    copySuccess.value = true
    showToast('已复制到剪贴板', 'mdi:content-copy')
    setTimeout(() => {
      copySuccess.value = false
    }, 2000)
  })
}

const saveToHistory = () => {
  const newItem = {
    from: from.value,
    to: to.value,
    amount: amount.value,
    result: result.value,
    rate: rateFixed.value,
    time: updateTime.value
  }

  history.value = [newItem, ...history.value.filter(
    h => !(h.from === newItem.from && h.to === newItem.to && h.amount === newItem.amount)
  )].slice(0, 6)

  localStorage.setItem('exchangeRateHistory', JSON.stringify(history.value))
}

const loadHistory = async (item) => {
  from.value = item.from
  to.value = item.to
  amount.value = item.amount
  showHistoryModal.value = false
  showToast('已加载历史记录', 'mdi:history')
  // 自动重新换算并定位
  await getRate()
}

const clearHistory = () => {
  history.value = []
  localStorage.removeItem('exchangeRateHistory')
  showHistoryModal.value = false
  showToast('历史记录已清空', 'mdi:trash-can')
}

// 生命周期
onMounted(() => {
  const savedHistory = localStorage.getItem('exchangeRateHistory')
  if (savedHistory) {
    history.value = JSON.parse(savedHistory)
  }
})

// 监听货币变化
watch([from, to], () => {
  result.value = ''
  rateFixed.value = ''
})
</script>

<style scoped>
.exchange-rate-container {
  max-width: 480px;
  margin: 0 auto;
  padding: 16px;
  padding-bottom: 100px;
}

/* 页面头部 */
.page-header {
  text-align: center;
  margin-bottom: 20px;
  padding-top: 8px;
}

.page-title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 6px 0;
}

.page-title svg {
  width: 28px;
  height: 28px;
  color: var(--primary-color);
}

.page-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
}

/* 主卡片 */
.main-card {
  background: var(--card-bg);
  border-radius: 20px;
  padding: 20px;
  box-shadow: var(--shadow);
  border: 1px solid var(--border-color);
  margin-bottom: 16px;
}

/* 货币选择器 */
.currency-selector {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.currency-box {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 12px;
  background: var(--input-bg);
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.currency-box:hover {
  border-color: var(--primary-color);
  background: var(--bg-secondary);
}

.currency-flag {
  font-size: 28px;
  line-height: 1;
}

.currency-info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.currency-code {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.currency-name {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.2;
}

.arrow-icon {
  width: 20px;
  height: 20px;
  color: var(--text-muted);
}

.swap-btn {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: none;
  background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s;
  flex-shrink: 0;
}

.swap-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(67, 97, 238, 0.4);
}

.swap-btn.swapping {
  transform: rotate(180deg);
}

.swap-btn svg {
  width: 24px;
  height: 24px;
}

/* 汇率信息条 */
.rate-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: linear-gradient(135deg, rgba(67, 97, 238, 0.08), rgba(67, 97, 238, 0.03));
  border-radius: 10px;
  margin-bottom: 16px;
}

.rate-label {
  font-size: 11px;
  color: var(--text-secondary);
}

.rate-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--primary-color);
}

/* 金额输入区 */
.amount-section {
  margin-bottom: 16px;
}

.amount-input-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px;
  background: var(--input-bg);
  border-radius: 14px;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.amount-input-box:focus-within {
  border-color: var(--primary-color);
  background: var(--bg-secondary);
}

.amount-symbol {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-secondary);
}

.amount-input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 32px;
  font-weight: 700;
  color: var(--text-primary);
  outline: none;
  width: 100%;
}

.amount-input::placeholder {
  color: var(--text-muted);
}

.clear-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.clear-btn:hover {
  color: var(--text-secondary);
}

.clear-btn svg {
  width: 20px;
  height: 20px;
}

.amount-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 6px;
  padding-left: 4px;
}

/* 快捷金额 */
.quick-amounts {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.quick-amount-btn {
  flex: 1;
  min-width: 60px;
  padding: 10px 8px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: var(--btn-secondary-bg);
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-amount-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.quick-amount-btn.active {
  background: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

/* 换算按钮 */
.convert-btn {
  width: 100%;
  padding: 16px;
  border: none;
  border-radius: 14px;
  background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
  color: white;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.2s;
}

.convert-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(67, 97, 238, 0.4);
}

.convert-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.convert-btn svg {
  width: 20px;
  height: 20px;
}

/* 结果卡片 */
.result-card {
  background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
  border-radius: 20px;
  padding: 20px;
  color: white;
  margin-bottom: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.result-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(67, 97, 238, 0.3);
}

.result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.result-label {
  font-size: 13px;
  opacity: 0.9;
}

.copy-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.copy-btn.copied {
  background: #4caf50;
}

.copy-btn svg {
  width: 18px;
  height: 18px;
}

.result-body {
  text-align: center;
}

.result-from,
.result-to {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 8px;
}

.result-amount {
  font-size: 24px;
  font-weight: 600;
  opacity: 0.9;
}

.result-amount.highlight {
  font-size: 36px;
  font-weight: 700;
  opacity: 1;
}

.result-currency {
  font-size: 16px;
  opacity: 0.8;
}

.result-equals {
  margin: 12px 0;
  opacity: 0.6;
}

.result-equals svg {
  width: 24px;
  height: 24px;
}

.result-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.2);
  font-size: 12px;
  opacity: 0.8;
}

.result-footer svg {
  width: 14px;
  height: 14px;
}

/* 损耗卡片 */
.loss-card {
  background: var(--card-bg);
  border-radius: 20px;
  padding: 16px 20px;
  box-shadow: var(--shadow);
  border: 1px solid var(--border-color);
  margin-bottom: 16px;
}

.loss-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  padding: 4px 0;
}

.loss-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.loss-title svg {
  width: 20px;
  height: 20px;
  color: var(--primary-color);
}

.toggle-icon {
  width: 20px;
  height: 20px;
  color: var(--text-muted);
  transition: transform 0.2s;
}

.loss-body {
  padding-top: 16px;
}

.loss-input-box {
  margin-bottom: 12px;
}

.loss-input-box label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.loss-input-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  background: var(--input-bg);
  border-radius: 12px;
  border: 1px solid var(--border-color);
}

.loss-symbol {
  font-size: 18px;
  color: var(--text-secondary);
}

.loss-input-wrapper input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  outline: none;
}

.calc-loss-btn {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--primary-color);
  border-radius: 12px;
  background: transparent;
  color: var(--primary-color);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-bottom: 16px;
  transition: all 0.2s;
}

.calc-loss-btn:hover:not(:disabled) {
  background: var(--primary-color);
  color: white;
}

.calc-loss-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.calc-loss-btn svg {
  width: 18px;
  height: 18px;
}

.loss-result {
  background: var(--bg-secondary);
  border-radius: 14px;
  padding: 16px;
}

.loss-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.loss-item {
  text-align: center;
}

.loss-item-label {
  display: block;
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.loss-item-value {
  display: block;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.loss-item.highlight .loss-item-value {
  color: #f44336;
  font-size: 18px;
}

/* 历史记录卡片 */
.history-card {
  background: var(--card-bg);
  border-radius: 20px;
  padding: 16px 20px;
  box-shadow: var(--shadow);
  border: 1px solid var(--border-color);
}

.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.history-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.history-title svg {
  width: 20px;
  height: 20px;
  color: var(--primary-color);
}

.clear-history-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.clear-history-btn:hover {
  background: var(--bg-secondary);
  color: var(--text-secondary);
}

.clear-history-btn svg {
  width: 18px;
  height: 18px;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.history-item:hover {
  background: var(--input-bg);
}

.history-currency-pair {
  display: flex;
  align-items: center;
  gap: 6px;
}

.history-flag {
  font-size: 18px;
}

.history-code {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.history-arrow {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
}

.history-details {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.history-amount {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.history-rate {
  font-size: 11px;
  color: var(--text-secondary);
}

/* FAB 历史记录按钮 */
.fab-history-btn {
  position: fixed;
  bottom: 90px;
  right: 20px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
  color: white;
  border: none;
  box-shadow: 0 4px 16px rgba(67, 97, 238, 0.4);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  transition: all 0.3s ease;
}

.fab-history-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 6px 20px rgba(67, 97, 238, 0.5);
}

.fab-history-btn svg {
  width: 24px;
  height: 24px;
}

.fab-badge {
  position: absolute;
  top: -2px;
  right: -2px;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  background: #f44336;
  color: white;
  font-size: 12px;
  font-weight: 600;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--bg-primary);
}

/* 历史记录弹窗 */
.history-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  backdrop-filter: blur(4px);
}

.history-modal {
  background: var(--card-bg);
  width: 100%;
  max-width: 480px;
  max-height: 70vh;
  border-radius: 20px 20px 0 0;
  display: flex;
  flex-direction: column;
  animation: slideUp 0.3s ease;
}

.history-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.history-modal-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.history-modal-header h3 svg {
  width: 20px;
  height: 20px;
  color: var(--primary-color);
}

.history-modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
}

.history-modal-body .history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-modal-body .history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: var(--bg-secondary);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.history-modal-body .history-item:hover {
  background: var(--input-bg);
  transform: translateX(4px);
}

.history-modal-footer {
  padding: 12px 16px 20px;
  border-top: 1px solid var(--border-color);
}

.history-modal-footer .clear-history-btn {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s;
}

.history-modal-footer .clear-history-btn:hover {
  background: rgba(244, 67, 54, 0.1);
  color: #f44336;
  border-color: #f44336;
}

.history-modal-footer .clear-history-btn svg {
  width: 18px;
  height: 18px;
}

/* 货币选择弹窗 */
.currency-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  backdrop-filter: blur(4px);
}

.currency-modal {
  background: var(--card-bg);
  width: 100%;
  max-width: 480px;
  max-height: 80vh;
  border-radius: 20px 20px 0 0;
  display: flex;
  flex-direction: column;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}

.currency-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.currency-modal-header h3 {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.close-btn:hover {
  background: var(--bg-secondary);
}

.close-btn svg {
  width: 20px;
  height: 20px;
}

.currency-search {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  margin: 12px 16px;
  background: var(--input-bg);
  border-radius: 12px;
  border: 1px solid var(--border-color);
}

.currency-search svg {
  width: 20px;
  height: 20px;
  color: var(--text-muted);
}

.currency-search input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 15px;
  color: var(--text-primary);
  outline: none;
}

.currency-search input::placeholder {
  color: var(--text-muted);
}

.currency-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 16px 20px;
}

.currency-group {
  margin-bottom: 16px;
}

.group-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  padding: 8px 4px;
  margin-bottom: 4px;
}

.currency-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.currency-option:hover {
  background: var(--bg-secondary);
}

.currency-option.active {
  background: rgba(67, 97, 238, 0.1);
}

.currency-option .currency-flag {
  font-size: 24px;
}

.currency-option .currency-info {
  flex: 1;
}

.currency-option .currency-code {
  font-size: 15px;
}

.currency-option .currency-name {
  font-size: 13px;
}

.check-icon {
  width: 20px;
  height: 20px;
  color: var(--primary-color);
}

/* Toast */
.toast {
  position: fixed;
  bottom: 100px;
  left: 50%;
  transform: translateX(-50%) translateY(100px);
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 12px 24px;
  border-radius: 24px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  opacity: 0;
  transition: all 0.3s;
  z-index: 2000;
  pointer-events: none;
}

.toast.show {
  transform: translateX(-50%) translateY(0);
  opacity: 1;
}

.toast svg {
  width: 18px;
  height: 18px;
}

/* 动画 */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
  max-height: 400px;
  opacity: 1;
}

.slide-enter-from,
.slide-leave-to {
  max-height: 0;
  opacity: 0;
  overflow: hidden;
}

/* 深色模式适配 */
.dark .result-card {
  background: linear-gradient(135deg, #5a7bf7, #4a5fd9);
}

.dark .rate-bar {
  background: linear-gradient(135deg, rgba(90, 123, 247, 0.15), rgba(90, 123, 247, 0.05));
}

/* 响应式 */
@media (max-width: 480px) {
  .exchange-rate-container {
    padding: 12px;
    padding-bottom: 90px;
  }

  .page-title {
    font-size: 20px;
  }

  .main-card,
  .result-card,
  .loss-card,
  .history-card {
    padding: 16px;
    border-radius: 16px;
  }

  .currency-flag {
    font-size: 24px;
  }

  .currency-code {
    font-size: 16px;
  }

  .amount-input {
    font-size: 28px;
  }

  .result-amount.highlight {
    font-size: 28px;
  }

  .quick-amount-btn {
    padding: 8px 6px;
    font-size: 12px;
  }

  .loss-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .loss-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
}
</style>
