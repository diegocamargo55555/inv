export function formatCurrency(amount: number | string, currency: string = 'BRL'): string {
  const numericAmount = typeof amount === 'string' ? parseFloat(amount) : amount
  if (isNaN(numericAmount)) return currency === 'BRL' ? 'R$ 0,00' : '$0.00'

  if (currency === 'BRL') {
    const isNegative = numericAmount < 0
    const absVal = Math.abs(numericAmount)
    const formatted = new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(absVal).replace(/\u00a0/g, ' ')

    return isNegative ? `-${formatted}` : formatted
  }

  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
  }).format(numericAmount)
}

export function formatPercentage(value: number | string): string {
  const num = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(num)) return '0.00%'

  const formatted = Math.abs(num).toFixed(2)
  if (num > 0) return `+${formatted}%`
  if (num < 0) return `-${formatted}%`
  return `${formatted}%`
}

export function formatInvoiceMonth(monthYear: string): string {
  // e.g. "2026-03" -> "Mar/2026"
  const parts = monthYear.split('-')
  if (parts.length !== 2) return monthYear

  const year = parts[0]
  const monthNum = parseInt(parts[1], 10)
  const monthNames = [
    'Jan', 'Fev', 'Mar', 'Abr', 'Mai', 'Jun',
    'Jul', 'Ago', 'Set', 'Out', 'Nov', 'Dez'
  ]

  if (monthNum >= 1 && monthNum <= 12) {
    return `${monthNames[monthNum - 1]}/${year}`
  }
  return monthYear
}
