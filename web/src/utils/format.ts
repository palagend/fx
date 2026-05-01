export const formatAmount = (amount: number | undefined): string => {
  if (amount === undefined) return '0.0000'
  return amount.toLocaleString('en-US', {
    minimumFractionDigits: 4,
    maximumFractionDigits: 4
  })
}

export const formatCompactAmount = (amount: number | undefined): string => {
  if (amount === undefined) return '$0'
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

export const formatDateTime = (date: string | number | undefined): string => {
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

export const formatDate = (date: string | number | undefined): string => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

export const formatPercent = (value: number | undefined, decimals: number = 2): string => {
  if (value === undefined) return '0%'
  const sign = value >= 0 ? '+' : '-'
  return sign + Math.abs(value).toFixed(decimals) + '%'
}

export const getChangeClass = (value: number): string => {
  if (value > 0) return 'positive'
  if (value < 0) return 'negative'
  return 'neutral'
}