package pwd

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
	"sync"
)

var agc *Argon2Configs
var rwm = sync.RWMutex{}

// Default Argon2Configs
const (
	MEMORY      uint32 = 64 * 1024
	ITER        uint32 = 3
	PARALLELISM uint8  = 2
	SALT_LEN    uint32 = 16
	KEY_LEN     uint32 = 32
)

type Argon2Configs struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func InitArgon2Configs(memory, iterations uint32, parallelism uint8, saltLength, keyLength uint32) *Argon2Configs {
	agc = &Argon2Configs{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		SaltLength:  saltLength,
		KeyLength:   keyLength,
	}
	return agc
}

func SetArgon2Configs(memory, iterations uint32, parallelism uint8, saltLength, keyLength uint32) *Argon2Configs {
	rwm.Lock()
	defer rwm.Unlock()

	// Set minimal requirements for Argon2
	if memory <= 0 {
		memory = MEMORY
	}
	if iterations <= 0 {
		iterations = ITER
	}
	if parallelism <= 0 {
		parallelism = PARALLELISM
	}
	if saltLength <= 0 {
		saltLength = SALT_LEN
	}
	if keyLength <= 0 {
		keyLength = KEY_LEN
	}

	// Re-config the itrlog
	agc = InitArgon2Configs(memory, iterations, parallelism, saltLength, keyLength)
	return agc
}

func init() {
	agc = InitArgon2Configs(MEMORY, ITER, PARALLELISM, SALT_LEN, KEY_LEN)
}

func HashAndSalt(password string) (string, error) {
	saltBytes := make([]byte, agc.SaltLength)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return "", err
	}

	argon2Hash := argon2.IDKey([]byte(password), saltBytes, agc.Iterations, agc.Memory, agc.Parallelism, agc.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(saltBytes)
	b64Argon2Hash := base64.RawStdEncoding.EncodeToString(argon2Hash)

	hash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, agc.Memory, agc.Iterations, agc.Parallelism, b64Salt, b64Argon2Hash)

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	// Decode the hash
	salt, key, err := DecodeHashPassword(hash)
	if err != nil {
		return false, err
	}

	// Hash the plain password
	argon2Hash := argon2.IDKey([]byte(password), salt, agc.Iterations, agc.Memory, agc.Parallelism, agc.KeyLength)

	// Compare the hashes
	if subtle.ConstantTimeCompare(key, argon2Hash) == 1 {
		return true, nil
	}

	return false, nil
}

func DecodeHashPassword(hash string) ([]byte, []byte, error) {
	// Check the format
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return nil, nil, errors.New("invalid hash format")
	}

	// Check the version
	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, err
	}

	// Compare the version
	if version != argon2.Version {
		vErr := fmt.Sprintf("invalid hash version. expected %d, got %d", argon2.Version, version)
		return nil, nil, errors.New(vErr)
	}

	// Decode the parameters
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &agc.Memory, &agc.Iterations, &agc.Parallelism)
	if err != nil {
		return nil, nil, err
	}

	// Decode the salt
	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, err
	}

	// Decode the key
	key, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, err
	}

	return salt, key, nil
}
