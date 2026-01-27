import { describe, it, expect } from 'vitest'
import { getErrorMessage } from './errors'

describe('errors', () => {
  describe('getErrorMessage', () => {
    it('should return error message from Error instance', () => {
      const error = new Error('Something went wrong')
      expect(getErrorMessage(error)).toBe('Something went wrong')
    })

    it('should return default fallback for non-Error values', () => {
      expect(getErrorMessage('string error')).toBe('Request failed')
      expect(getErrorMessage(123)).toBe('Request failed')
      expect(getErrorMessage(null)).toBe('Request failed')
      expect(getErrorMessage(undefined)).toBe('Request failed')
      expect(getErrorMessage({ message: 'object' })).toBe('Request failed')
    })

    it('should return custom fallback when provided', () => {
      expect(getErrorMessage('string error', 'Custom fallback')).toBe('Custom fallback')
      expect(getErrorMessage(null, 'Network error')).toBe('Network error')
    })
  })
})
