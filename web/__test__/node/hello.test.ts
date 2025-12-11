import { describe, it, expect } from 'vitest'

describe('Node.js Hello World', () => {
  it('should pass a simple assertion', () => {
    expect(1 + 1).toBe(2)
  })

  it('should work with strings', () => {
    const greeting = 'Hello World'
    expect(greeting).toBe('Hello World')
  })

  it('should work with objects', () => {
    const obj = { name: 'test', value: 42 }
    expect(obj).toEqual({ name: 'test', value: 42 })
  })
})
