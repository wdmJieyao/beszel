package hub

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	configBackupCryptoAlgorithm = "xchacha20poly1305"
	configBackupCryptoKDF       = "argon2id"
)

func encryptConfigBackupSecret(plaintext string, credential string, contentType string) (*ConfigBackupSecret, error) {
	if credential == "" {
		return nil, fmt.Errorf("encryption credential is required")
	}
	salt := make([]byte, 16)
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	key := deriveConfigBackupKey(credential, salt)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), []byte(contentType))
	return &ConfigBackupSecret{
		Encrypted:   base64.StdEncoding.EncodeToString(ciphertext),
		Nonce:       base64.StdEncoding.EncodeToString(nonce),
		Salt:        base64.StdEncoding.EncodeToString(salt),
		Algorithm:   configBackupCryptoAlgorithm,
		KDF:         configBackupCryptoKDF,
		ContentType: contentType,
	}, nil
}

func decryptConfigBackupSecret(secret *ConfigBackupSecret, credential string) (string, error) {
	if secret == nil || secret.Redacted || secret.Encrypted == "" {
		return "", nil
	}
	if credential == "" {
		return "", fmt.Errorf("decryption credential is required")
	}
	if secret.Algorithm != "" && secret.Algorithm != configBackupCryptoAlgorithm {
		return "", fmt.Errorf("unsupported secret algorithm")
	}
	if secret.KDF != "" && secret.KDF != configBackupCryptoKDF {
		return "", fmt.Errorf("unsupported secret kdf")
	}
	salt, err := base64.StdEncoding.DecodeString(secret.Salt)
	if err != nil {
		return "", fmt.Errorf("invalid secret salt")
	}
	nonce, err := base64.StdEncoding.DecodeString(secret.Nonce)
	if err != nil {
		return "", fmt.Errorf("invalid secret nonce")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(secret.Encrypted)
	if err != nil {
		return "", fmt.Errorf("invalid secret ciphertext")
	}
	key := deriveConfigBackupKey(credential, salt)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return "", err
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, []byte(secret.ContentType))
	if err != nil {
		return "", fmt.Errorf("decryption failed")
	}
	return string(plaintext), nil
}

func deriveConfigBackupKey(credential string, salt []byte) []byte {
	return argon2.IDKey([]byte(credential), salt, 1, 64*1024, 4, chacha20poly1305.KeySize)
}

func redactedConfigBackupSecret(contentType string) *ConfigBackupSecret {
	return &ConfigBackupSecret{
		Redacted:    true,
		ContentType: contentType,
	}
}

func configBackupHasEncryptedSecrets(document ConfigBackupDocument) bool {
	var found bool
	visit := func(secret *ConfigBackupSecret) {
		if secret != nil && secret.Encrypted != "" {
			found = true
		}
	}
	for _, system := range document.Systems {
		visit(system.Token)
	}
	for _, settings := range document.Notifications.UserSettings {
		for i := range settings.Webhooks {
			visit(&settings.Webhooks[i])
		}
	}
	visit(document.Notifications.Telegram.Settings.BotToken)
	return found
}
