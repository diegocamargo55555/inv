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
  const [year, month] = monthYear.split('-')
  if (!year || !month) return monthYear

  const date = new Date(Date.UTC(Number(year), Number(month) - 1, 1))
  if (isNaN(date.getTime())) return monthYear

  const rawMonth = date.toLocaleDateString('pt-BR', { month: 'short', timeZone: 'UTC' })
  const cleanMonth = rawMonth.replace('.', '').slice(0, 3)
  const capitalized = cleanMonth.charAt(0).toUpperCase() + cleanMonth.slice(1)
  return `${capitalized}/${year}`
}
