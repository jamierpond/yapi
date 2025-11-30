import { describe, it, expect } from 'vitest'
import { zip, unzip } from '../../app/_lib/gzip'

describe('gzip module', () => {
  describe('zip', () => {
    it('should compress a string', () => {
      const input = 'Hello, World!'
      const compressed = zip(input)

      expect(compressed).toBeInstanceOf(Uint8Array)
      expect(compressed.length).toBeGreaterThan(0)
    })

    it('should compress a Uint8Array', () => {
      const input = new Uint8Array([72, 101, 108, 108, 111])
      const compressed = zip(input)

      expect(compressed).toBeInstanceOf(Uint8Array)
      expect(compressed.length).toBeGreaterThan(0)
    })

    it('should compress an empty string', () => {
      const input = ''
      const compressed = zip(input)

      expect(compressed).toBeInstanceOf(Uint8Array)
      expect(compressed.length).toBeGreaterThan(0) // gzip header is still present
    })

    it('should compress long strings more efficiently', () => {
      const shortText = 'Hello'
      const longRepeatingText = 'Hello'.repeat(1000)

      const compressedShort = zip(shortText)
      const compressedLong = zip(longRepeatingText)

      // Compression should be more efficient for repeating text
      // The ratio should be better for the long repeating text
      const shortRatio = compressedShort.length / shortText.length
      const longRatio = compressedLong.length / longRepeatingText.length

      expect(longRatio).toBeLessThan(shortRatio)
    })

    it('should handle unicode characters', () => {
      const input = '你好世界 🌍 مرحبا'
      const compressed = zip(input)

      expect(compressed).toBeInstanceOf(Uint8Array)
      expect(compressed.length).toBeGreaterThan(0)
    })

    it('should produce different outputs for different inputs', () => {
      const input1 = 'Hello'
      const input2 = 'World'

      const compressed1 = zip(input1)
      const compressed2 = zip(input2)

      expect(compressed1).not.toEqual(compressed2)
    })

    it('should handle binary data', () => {
      const input = new Uint8Array([0, 1, 2, 3, 4, 5, 255, 254, 253])
      const compressed = zip(input)

      expect(compressed).toBeInstanceOf(Uint8Array)
      expect(compressed.length).toBeGreaterThan(0)
    })
  })

  describe('unzip', () => {
    it('should decompress to string by default', () => {
      const input = 'Hello, World!'
      const compressed = zip(input)
      const decompressed = unzip(compressed)

      expect(typeof decompressed).toBe('string')
      expect(decompressed).toBe(input)
    })

    it('should decompress to string with explicit utf8 encoding', () => {
      const input = 'Hello, World!'
      const compressed = zip(input)
      const decompressed = unzip(compressed, 'utf8')

      expect(typeof decompressed).toBe('string')
      expect(decompressed).toBe(input)
    })

    it('should decompress to Uint8Array with binary encoding', () => {
      const input = 'Hello, World!'
      const compressed = zip(input)
      const decompressed = unzip(compressed, 'binary')

      expect(decompressed).toBeInstanceOf(Uint8Array)
      expect(new TextDecoder().decode(decompressed)).toBe(input)
    })

    it('should decompress an empty string', () => {
      const input = ''
      const compressed = zip(input)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe('')
    })

    it('should handle unicode characters', () => {
      const input = '你好世界 🌍 مرحبا'
      const compressed = zip(input)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(input)
    })

    it('should throw error on invalid gzip data', () => {
      const invalidData = new Uint8Array([1, 2, 3, 4, 5])

      expect(() => unzip(invalidData)).toThrow()
    })
  })

  describe('zip and unzip round-trip', () => {
    it('should correctly round-trip a simple string', () => {
      const original = 'Hello, World!'
      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(original)
    })

    it('should correctly round-trip a long string', () => {
      const original = 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. '.repeat(100)
      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(original)
    })

    it('should correctly round-trip unicode text', () => {
      const original = '🎉 Testing 测试 тест परीक्षण 🚀'
      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(original)
    })

    it('should correctly round-trip binary data', () => {
      const original = new Uint8Array([0, 1, 2, 3, 127, 128, 255])
      const compressed = zip(original)
      const decompressed = unzip(compressed, 'binary')

      expect(decompressed).toEqual(original)
    })

    it('should correctly round-trip empty data', () => {
      const original = ''
      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(original)
    })

    it('should correctly round-trip JSON data', () => {
      const obj = {
        name: 'test',
        value: 42,
        nested: { array: [1, 2, 3] }
      }
      const original = JSON.stringify(obj)
      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(JSON.parse(decompressed)).toEqual(obj)
    })

    it('should correctly round-trip newlines and special characters', () => {
      const original = 'Line 1\nLine 2\r\nLine 3\tTabbed\0Null'
      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(original)
    })
  })

  describe('compression efficiency', () => {
    it('should compress repetitive data efficiently', () => {
      const original = 'AAAAAAAAAA'.repeat(100)
      const compressed = zip(original)

      // Highly repetitive data should compress well
      expect(compressed.length).toBeLessThan(original.length)
    })

    it('should handle already compressed/random data', () => {
      // Random-looking data doesn't compress well
      const original = new Uint8Array(100).map(() => Math.floor(Math.random() * 256))
      const compressed = zip(original)

      // Compressed size might be larger due to gzip overhead
      expect(compressed).toBeInstanceOf(Uint8Array)
    })

    it('should show compression ratio for typical text', () => {
      const original = `
        The quick brown fox jumps over the lazy dog.
        This is a sample text that should compress reasonably well.
        Repetition helps with compression. Repetition helps with compression.
      `.repeat(10)

      const compressed = zip(original)
      const ratio = compressed.length / original.length

      // Text should compress to less than original
      expect(ratio).toBeLessThan(1)
      expect(compressed.length).toBeLessThan(original.length)
    })
  })

  describe('edge cases', () => {
    it('should handle very long strings', () => {
      const original = 'x'.repeat(100000)
      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(original)
      expect(compressed.length).toBeLessThan(original.length)
    })

    it('should handle single character', () => {
      const original = 'x'
      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(original)
    })

    it('should handle all printable ASCII characters', () => {
      let original = ''
      for (let i = 32; i <= 126; i++) {
        original += String.fromCharCode(i)
      }

      const compressed = zip(original)
      const decompressed = unzip(compressed)

      expect(decompressed).toBe(original)
    })

    it('should handle mixed string and binary round-trip', () => {
      const text = 'Hello World'
      const binary = new TextEncoder().encode(text)

      const compressedFromString = zip(text)
      const compressedFromBinary = zip(binary)

      const decompressedToString = unzip(compressedFromString)
      const decompressedToBinary = unzip(compressedFromBinary, 'binary')

      expect(decompressedToString).toBe(text)
      expect(new TextDecoder().decode(decompressedToBinary)).toBe(text)
    })

    it('should preserve exact byte values in binary mode', () => {
      const original = new Uint8Array([0, 1, 127, 128, 254, 255])
      const compressed = zip(original)
      const decompressed = unzip(compressed, 'binary')

      expect(Array.from(decompressed)).toEqual(Array.from(original))
    })
  })
})
