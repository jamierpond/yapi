import { describe, it, expect } from 'vitest'
import { yapiEncode, yapiDecode } from '../../app/_lib/yapi-encode'

describe('yapi-encode module', () => {
  describe('yapiEncode', () => {
    it('should encode a simple string', () => {
      const input = 'Hello, World!'
      const encoded = yapiEncode(input)

      expect(typeof encoded).toBe('string')
      expect(encoded.length).toBeGreaterThan(0)
    })

    it('should produce a URL-safe string', () => {
      const input = 'Testing URL safety with special chars: &=+/?#[]@!$\'()*,;'
      const encoded = yapiEncode(input)

      // Encoded string should not need URL encoding
      expect(encodeURIComponent(encoded)).toBe(encoded)
    })

    it('should encode an empty string', () => {
      const input = ''
      const encoded = yapiEncode(input)

      expect(typeof encoded).toBe('string')
      expect(encoded.length).toBeGreaterThan(0)
    })

    it('should encode unicode characters', () => {
      const input = '你好世界 🌍 مرحبا Привет'
      const encoded = yapiEncode(input)

      expect(typeof encoded).toBe('string')
      expect(encoded.length).toBeGreaterThan(0)
    })

    it('should produce different encodings for different inputs', () => {
      const input1 = 'State 1'
      const input2 = 'State 2'

      const encoded1 = yapiEncode(input1)
      const encoded2 = yapiEncode(input2)

      expect(encoded1).not.toBe(encoded2)
    })

    it('should handle JSON stringified objects', () => {
      const obj = { name: 'test', value: 42, nested: { array: [1, 2, 3] } }
      const input = JSON.stringify(obj)
      const encoded = yapiEncode(input)

      expect(typeof encoded).toBe('string')
      expect(encoded.length).toBeGreaterThan(0)
    })

    it('should handle long strings efficiently', () => {
      const longText = 'Lorem ipsum dolor sit amet. '.repeat(100)
      const encoded = yapiEncode(longText)

      expect(typeof encoded).toBe('string')
      // Due to compression, encoded should be shorter than original
      expect(encoded.length).toBeLessThan(longText.length)
    })

    it('should handle newlines and whitespace', () => {
      const input = 'Line 1\nLine 2\r\nLine 3\tTabbed'
      const encoded = yapiEncode(input)

      expect(typeof encoded).toBe('string')
      expect(encoded.length).toBeGreaterThan(0)
    })
  })

  describe('yapiDecode', () => {
    it('should decode an encoded string', () => {
      const original = 'Hello, World!'
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should decode an empty string', () => {
      const original = ''
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should decode unicode characters', () => {
      const original = '你好世界 🌍 مرحبا Привет'
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should throw error on invalid encoded string', () => {
      const invalid = 'invalid@string#with$bad%chars'

      expect(() => yapiDecode(invalid)).toThrow()
    })

    it('should throw error on corrupted data', () => {
      const original = 'Hello, World!'
      const encoded = yapiEncode(original)
      // Corrupt the encoded string by removing some characters
      const corrupted = encoded.slice(0, -5)

      expect(() => yapiDecode(corrupted)).toThrow()
    })
  })

  describe('yapiEncode and yapiDecode round-trip', () => {
    it('should correctly round-trip a simple string', () => {
      const original = 'Hello, World!'
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should correctly round-trip a long string', () => {
      const original = 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. '.repeat(100)
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should correctly round-trip unicode text', () => {
      const original = '🎉 Testing 测试 тест परीक्षण ทดสอบ 🚀'
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should correctly round-trip JSON data', () => {
      const obj = {
        name: 'YAPI State',
        version: '1.0.0',
        config: {
          endpoint: 'http://localhost:3000',
          timeout: 5000,
          headers: { 'Content-Type': 'application/json' }
        },
        items: [1, 2, 3, 4, 5]
      }
      const original = JSON.stringify(obj)
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(JSON.parse(decoded)).toEqual(obj)
    })

    it('should correctly round-trip empty string', () => {
      const original = ''
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should correctly round-trip special characters', () => {
      const original = 'Special chars: !@#$%^&*()_+-=[]{}|;:\'",.<>?/~`\n\r\t'
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should correctly round-trip multiline text', () => {
      const original = `Line 1
Line 2
Line 3
\tIndented line
\t\tDouble indented`
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should correctly round-trip very long text', () => {
      const original = 'x'.repeat(10000)
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })
  })

  describe('compression and URL safety', () => {
    it('should compress repetitive data efficiently', () => {
      const original = 'AAAA '.repeat(1000)
      const encoded = yapiEncode(original)

      // Should be much shorter due to compression
      expect(encoded.length).toBeLessThan(original.length)
    })

    it('should produce URL-safe output', () => {
      const testCases = [
        'Simple text',
        'Text with spaces and punctuation!',
        '{"json": "object"}',
        '你好世界',
        'Multi\nLine\nText'
      ]

      testCases.forEach(input => {
        const encoded = yapiEncode(input)
        // URL-safe means encodeURIComponent doesn't change it
        expect(encodeURIComponent(encoded)).toBe(encoded)
      })
    })

    it('should handle typical YAPI state objects', () => {
      const yapiState = {
        endpoint: 'http://api.example.com/v1/users',
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer token123'
        },
        body: {
          name: 'John Doe',
          email: 'john@example.com',
          preferences: {
            theme: 'dark',
            notifications: true
          }
        }
      }

      const original = JSON.stringify(yapiState, null, 2)
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(JSON.parse(decoded)).toEqual(yapiState)
      expect(encodeURIComponent(encoded)).toBe(encoded)
    })

    it('should show compression benefits for structured data', () => {
      const structuredData = JSON.stringify({
        users: Array(100).fill({
          id: 1,
          name: 'User Name',
          email: 'user@example.com',
          active: true
        })
      })

      const encoded = yapiEncode(structuredData)

      // Compression should reduce size significantly
      expect(encoded.length).toBeLessThan(structuredData.length * 0.5)
    })
  })

  describe('edge cases', () => {
    it('should handle single character', () => {
      const original = 'x'
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should handle only whitespace', () => {
      const original = '   \n\t\r\n   '
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should handle all printable ASCII characters', () => {
      let original = ''
      for (let i = 32; i <= 126; i++) {
        original += String.fromCharCode(i)
      }

      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should handle emoji sequences', () => {
      const original = '👨‍👩‍👧‍👦 👍🏽 🏳️‍🌈'
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })

    it('should produce consistent output for same input', () => {
      const original = 'Consistent test'
      const encoded1 = yapiEncode(original)
      const encoded2 = yapiEncode(original)

      expect(encoded1).toBe(encoded2)
    })

    it('should handle YAML-like content', () => {
      const original = `
openapi: "3.0.0"
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      summary: Test endpoint
      responses:
        '200':
          description: Success
`
      const encoded = yapiEncode(original)
      const decoded = yapiDecode(encoded)

      expect(decoded).toBe(original)
    })
  })

  describe('integration with URL usage', () => {
    it('should work in URL path context', () => {
      const state = JSON.stringify({ endpoint: 'http://api.test.com', method: 'GET' })
      const encoded = yapiEncode(state)

      // Simulate URL usage
      const url = `https://yapi.run/c/${encoded}`
      const extractedEncoded = url.split('/c/')[1]

      const decoded = yapiDecode(extractedEncoded)
      expect(JSON.parse(decoded)).toEqual(JSON.parse(state))
    })

    it('should work in query parameter context', () => {
      const state = JSON.stringify({ test: 'value' })
      const encoded = yapiEncode(state)

      // Simulate query parameter
      const url = `https://example.com?state=${encoded}`
      const extractedEncoded = new URL(url).searchParams.get('state')

      const decoded = yapiDecode(extractedEncoded!)
      expect(decoded).toBe(state)
    })
  })
})
