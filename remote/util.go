package remote

import (
	"crypto/aes"
	"fmt"
)

// TODO: is this ideomatic or efficient?
func PadPKCS7(input []byte, bs int) []byte {
	output := input[:]
	padSize := uint8(bs - len(input)%bs)
	for i := uint8(0); i < padSize; i++ {
		output = append(output, padSize)
	}
	return output
}

func AESCBC(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("Failed to create cipher %s", err)
	}

	bs := block.BlockSize()
	paddedPlaintext := PadPKCS7(plaintext, bs)

	ciphertext := make([]byte, len(paddedPlaintext))
	// TODO: this can be cleaner
	tmp := ciphertext
	for len(paddedPlaintext) > 0 {
		block.Encrypt(tmp, paddedPlaintext)
		paddedPlaintext = paddedPlaintext[bs:]
		tmp = tmp[bs:]
	}
	return ciphertext, nil
}
