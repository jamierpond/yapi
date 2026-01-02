import { gzip as gzipCompress, ungzip as gzipDecompress } from 'pako'

/**
 * Compresses data using gzip compression
 * @param data - The data to compress (string or Uint8Array)
 * @returns Compressed data as Uint8Array
 */
export function zip(data: string | Uint8Array): Uint8Array {
  const input = typeof data === 'string' ? new TextEncoder().encode(data) : data
  return gzipCompress(input)
}

/**
 * Decompresses gzip-compressed data
 * @param data - The compressed data as Uint8Array
 * @param encoding - Optional encoding for output ('utf8' or 'binary'). Default: 'utf8'
 * @returns Decompressed data as string (if encoding is 'utf8') or Uint8Array (if encoding is 'binary')
 */
export function unzip(data: Uint8Array, encoding?: 'utf8'): string
export function unzip(data: Uint8Array, encoding: 'binary'): Uint8Array
export function unzip(data: Uint8Array, encoding: 'utf8' | 'binary' = 'utf8'): string | Uint8Array {
  const decompressed = gzipDecompress(data)

  if (encoding === 'binary') {
    return decompressed
  }

  return new TextDecoder().decode(decompressed)
}
