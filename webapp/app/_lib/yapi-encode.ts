import { zip, unzip } from "./gzip";
import { encodeBuffer, decodeToBuffer } from "./encoding";

export function yapiEncode(state: string): string {
  const utf8Encoder = new TextEncoder();
  const utf8Data = utf8Encoder.encode(state);
  const compressedData = zip(utf8Data);
  const encodedString = encodeBuffer(compressedData);
  return encodedString;
}

export function yapiDecode(encodedState: string): string {
  const compressedData = decodeToBuffer(encodedState);
  const utf8Data = unzip(compressedData);
  const utf8Decoder = new TextDecoder();
  const state = utf8Decoder.decode(utf8Data);
  return state;
}
