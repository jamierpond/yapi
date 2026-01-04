// Package share provides URL encoding for sharing configs.
package share

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"math/big"
)

var characterSet = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~")
var base = big.NewInt(int64(len(characterSet)))

func encodeBuffer(data []byte) string {
	value := new(big.Int).SetBytes(data)

	if value.Sign() == 0 {
		return ""
	}

	var encoded bytes.Buffer
	zero := big.NewInt(0)
	mod := new(big.Int)

	for value.Cmp(zero) > 0 {
		value.DivMod(value, base, mod)
		encoded.WriteByte(characterSet[mod.Int64()])
	}

	// Reverse the result
	result := encoded.Bytes()
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encode compresses and encodes content for sharing via yapi.run/c/{encoded}
func Encode(content string) (string, error) {
	compressed, err := gzipCompress([]byte(content))
	if err != nil {
		return "", fmt.Errorf("compression failed: %w", err)
	}
	return encodeBuffer(compressed), nil
}
