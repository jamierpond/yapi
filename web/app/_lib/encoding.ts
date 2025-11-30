// this route is for yapi.run/c/{encoded_state}
// const safe = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~";

export const CHARACTER_SET = [
  "A","B","C","D","E","F","G","H","I","J","K","L","M",
  "N","O","P","Q","R","S","T","U","V","W","X","Y","Z",
  "a","b","c","d","e","f","g","h","i","j","k","l","m",
  "n","o","p","q","r","s","t","u","v","w","x","y","z",
  "0","1","2","3","4","5","6","7","8","9",
  "-","_",".","~"
] as const;

export const BASE = CHARACTER_SET.length;
export type EncodedChar = typeof CHARACTER_SET[number];

export function encodeBuffer(buffer: Uint8Array): string {
  let value = BigInt(0);
  for (let i = 0; i < buffer.length; i++) {
    value = (value << BigInt(8)) + BigInt(buffer[i]);
  }

  let encoded = "";
  while (value > 0) {
    const remainder = Number(value % BigInt(BASE));
    encoded = CHARACTER_SET[remainder] + encoded;
    value = value / BigInt(BASE);
  }

  return encoded;
}

export function decodeToBuffer(encoded: string): Uint8Array {
  let value = BigInt(0);
  for (let i = 0; i < encoded.length; i++) {
    const index = CHARACTER_SET.indexOf(encoded[i] as EncodedChar);
    if (index === -1) {
      throw new Error(`Invalid character '${encoded[i]}' in encoded string.`);
    }
    value = value * BigInt(BASE) + BigInt(index);
  }

  const bytes: number[] = [];
  while (value > 0) {
    bytes.push(Number(value & BigInt(0xFF)));
    value = value >> BigInt(8);
  }

  return new Uint8Array(bytes.reverse());
}

