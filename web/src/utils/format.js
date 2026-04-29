// 格式化金额（完整格式，保留4位小数）
export const formatAmount = (amount) => {
  if (!amount && amount !== 0) return '0.0000'
  return amount.toLocaleString('en-US', {
    minimumFractionDigits: 4,
    maximumFractionDigits: 4
  })
}

// 格式化大金额为缩写形式（K/M/B）
export const formatCompactAmount = (amount) => {
  if (!amount && amount !== 0) return '$0'
  const absAmount = Math.abs(amount)
  if (absAmount >= 1e9) {
    return '$' + (amount / 1e9).toFixed(2) + 'B'
  } else if (absAmount >= 1e6) {
    return '$' + (amount / 1e6).toFixed(2) + 'M'
  } else if (absAmount >= 1e3) {
    return '$' + (amount / 1e3).toFixed(2) + 'K'
  }
  return '$' + amount.toFixed(2)
}

// 格式化日期时间
export const formatDateTime = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 格式化日期（仅日期）
export const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

// 格式化百分比
export const formatPercent = (value, decimals = 2) => {
  if (!value && value !== 0) return '0%'
  const sign = value >= 0 ? '+' : '-'
  return sign + Math.abs(value).toFixed(decimals) + '%'
}

// 获取变化样式类名
export const getChangeClass = (value) => {
  if (value > 0) return 'positive'
  if (value < 0) return 'negative'
  return 'neutral'
}
