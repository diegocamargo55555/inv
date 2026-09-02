import { describe, it, expect } from 'vitest'
import { formatCurrency, formatPercentage, formatInvoiceMonth } from './formatters'

describe('Formatters Utils (TDD)', () => {
  it('formats BRL currency properly', () => {
    expect(formatCurrency(1250.5, 'BRL')).toBe('R$ 1.250,50')
    expect(formatCurrency(0, 'BRL')).toBe('R$ 0,00')
    expect(formatCurrency(-50.25, 'BRL')).toBe('-R$ 50,25')
  })

  it('formats USD currency properly', () => {
    expect(formatCurrency(1250.5, 'USD')).toBe('$1,250.50')
    expect(formatCurrency(0, 'USD')).toBe('$0.00')
  })

  it('formats percentage with sign and decimal places', () => {
    expect(formatPercentage(12.456)).toBe('+12.46%')
    expect(formatPercentage(-3.2)).toBe('-3.20%')
    expect(formatPercentage(0)).toBe('0.00%')
  })

  it('formats invoice month year to readable label', () => {
    expect(formatInvoiceMonth('2026-03')).toBe('Mar/2026')
    expect(formatInvoiceMonth('2026-12')).toBe('Dez/2026')
  })
})
