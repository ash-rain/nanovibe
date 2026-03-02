package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
)

// machineKey derives a deterministic 32-byte key from the machine's hostname
// and primary MAC address. The key is never stored or logged.
func machineKey() []byte {
	hostname, _ := os.Hostname()
	mac := primaryMAC()
	raw := sha256.Sum256([]byte(hostname + ":" + mac))
	return raw[:]
}

// primaryMAC returns the hardware address of the first non-loopback interface
// that has a MAC address. Returns empty string on failure.
func primaryMAC() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		return iface.HardwareAddr.String()
	}
	return ""
}

// Encrypt encrypts plaintext using AES-256-GCM with the machine key.
// The output is base64-encoded; the random 12-byte nonce is prepended to the ciphertext.
func Encrypt(plaintext string) (string, error) {
	key := machineKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("crypto: new cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: new gcm: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext (with prepended nonce) using the machine key.
func Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("crypto: base64 decode: %w", err)
	}

	key := machineKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("crypto: new cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: new gcm: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("crypto: ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: decrypt: %w", err)
	}

	return string(plaintext), nil
}
