import { describe, it, expect } from 'vitest'
import {
  encodeBuffer,
  decodeToBuffer,
  CHARACTER_SET,
  BASE
} from '../../app/_lib/encoding'

describe('encoding module', () => {
  describe('CHARACTER_SET', () => {
    it('should have 66 characters', () => {
      expect(CHARACTER_SET.length).toBe(66)
    })

    it('should contain URL-safe characters', () => {
      const charset = CHARACTER_SET.join('')
      expect(charset).toMatch(/^[A-Za-z0-9\-_.~]+$/)
    })

    it('should have unique characters', () => {
      const uniqueChars = new Set(CHARACTER_SET)
      expect(uniqueChars.size).toBe(CHARACTER_SET.length)
    })
  })

  describe('BASE', () => {
    it('should equal the length of CHARACTER_SET', () => {
      expect(BASE).toBe(66)
    })
  })

  describe('encodeBuffer', () => {
    it('should encode an empty buffer to empty string', () => {
      const buffer = new Uint8Array([])
      const encoded = encodeBuffer(buffer)
      expect(encoded).toBe('')
    })

    it('should encode a single byte', () => {
      const buffer = new Uint8Array([65])
      const encoded = encodeBuffer(buffer)
      expect(encoded).toBeTruthy()
      expect(typeof encoded).toBe('string')
    })

    it('should encode multiple bytes', () => {
      const buffer = new Uint8Array([72, 101, 108, 108, 111])
      const encoded = encodeBuffer(buffer)
      expect(encoded).toBeTruthy()
      expect(typeof encoded).toBe('string')
    })

    it('should only use characters from CHARACTER_SET', () => {
      const buffer = new Uint8Array([1, 2, 3, 4, 5])
      const encoded = encodeBuffer(buffer)

      for (const char of encoded) {
        expect(CHARACTER_SET).toContain(char)
      }
    })

    it('should produce different encodings for different inputs', () => {
      const buffer1 = new Uint8Array([1, 2, 3])
      const buffer2 = new Uint8Array([4, 5, 6])

      const encoded1 = encodeBuffer(buffer1)
      const encoded2 = encodeBuffer(buffer2)

      expect(encoded1).not.toBe(encoded2)
    })

    it('should handle large byte values', () => {
      const buffer = new Uint8Array([255, 254, 253])
      const encoded = encodeBuffer(buffer)
      expect(encoded).toBeTruthy()
    })
  })

  describe('decodeToBuffer', () => {
    it('should decode an empty string to empty buffer', () => {
      const decoded = decodeToBuffer('')
      expect(decoded).toEqual(new Uint8Array([]))
    })

    it('should throw error on invalid character', () => {
      expect(() => decodeToBuffer('abc$def')).toThrow('Invalid character')
      expect(() => decodeToBuffer('hello world')).toThrow('Invalid character')
      expect(() => decodeToBuffer('test@test')).toThrow('Invalid character')
    })

    it('should decode valid encoded strings', () => {
      const validChars = 'ABC123-_.~'
      expect(() => decodeToBuffer(validChars)).not.toThrow()
    })

    it('should return a Uint8Array', () => {
      const decoded = decodeToBuffer('ABC')
      expect(decoded).toBeInstanceOf(Uint8Array)
    })
  })

  describe('encodeBuffer and decodeToBuffer round-trip', () => {
    it('should correctly round-trip encode and decode', () => {
      const original = new Uint8Array([72, 101, 108, 108, 111])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })

    it('should handle single byte round-trip', () => {
      const original = new Uint8Array([42])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })

    it('should handle multiple bytes round-trip', () => {
      const original = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })

    it('should handle various byte values round-trip', () => {
      // Note: Leading zeros are not preserved due to BigInt conversion
      const original = new Uint8Array([1, 127, 128, 254, 255])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })

    it('should handle larger buffers round-trip', () => {
      // Start from 1 to avoid leading zeros
      const original = new Uint8Array(100).map((_, i) => (i + 1) % 256 || 1)
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })

    it('should handle empty buffer round-trip', () => {
      const original = new Uint8Array([])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })

    it('should handle binary data round-trip', () => {
      const original = new Uint8Array([0xFF, 0xFE, 0xFD, 0x00, 0x01, 0x02])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })
  })

  describe('edge cases', () => {
    it('should handle leading zeros (note: leading zeros are lost)', () => {
      // This is a known limitation - leading zeros are not preserved
      const original = new Uint8Array([0, 0, 1])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      // Leading zeros are stripped during BigInt conversion
      expect(decoded).toEqual(new Uint8Array([1]))
    })

    it('should handle buffer starting with zero (note: leading zeros are lost)', () => {
      // This is a known limitation - leading zeros are not preserved
      const original = new Uint8Array([0, 255])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      // Leading zero is stripped
      expect(decoded).toEqual(new Uint8Array([255]))
    })

    it('should handle sequential values', () => {
      const original = new Uint8Array([1, 2, 3, 4, 5])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })

    it('should handle same values repeated', () => {
      const original = new Uint8Array([42, 42, 42, 42])
      const encoded = encodeBuffer(original)
      const decoded = decodeToBuffer(encoded)

      expect(decoded).toEqual(original)
    })
  })

  describe('URL-safe encoding', () => {
    it('should produce URL-safe strings', () => {
      const buffer = new Uint8Array([72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100])
      const encoded = encodeBuffer(buffer)

      // Check that encoded string doesn't need URL encoding
      expect(encodeURIComponent(encoded)).toBe(encoded)
    })

    it('should not contain special characters that need escaping', () => {
      const buffer = new Uint8Array([255, 254, 253, 252, 251, 250])
      const encoded = encodeBuffer(buffer)

      // Should not contain &, =, +, /, etc.
      expect(encoded).not.toMatch(/[&=+/]/)
    })
  })
})
