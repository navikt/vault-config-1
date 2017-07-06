package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	wrappedCipherRegex = regexp.MustCompile(`@encrypted_data\((.*)\)`)
	wrappedHmacRegex   = regexp.MustCompile(`@hmac\((.*)\)`)
)

// EncryptionObject contains all the variables and methods
// associated with encrypting and decrypting data
type EncryptionObject struct {
	Key         []byte
	CipherText  []byte
	PlainText   []byte
	HMAC        []byte
	WrappedData string
}

// Encrypt will crypto data with specified key
func (e *EncryptionObject) Encrypt() error {
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return fmt.Errorf("Error creating AES block: %v", err)
	}

	e.CipherText = make([]byte, aes.BlockSize+len(e.PlainText))
	iv := e.CipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return fmt.Errorf("Error creating iv: %v", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(e.CipherText[aes.BlockSize:], e.PlainText)
	e.HMAC, err = CreateHMAC(e.Key, e.CipherText)
	if err != nil {
		return fmt.Errorf("Creating HMAC: %v", err)
	}
	e.WrapCrypto()

	return nil
}

// Decrypt will decrypt data with specified key
func (e *EncryptionObject) Decrypt() error {
	h, _ := CreateHMAC(e.Key, e.CipherText)
	if !hmac.Equal(e.HMAC, h) {
		return fmt.Errorf("HMAC failure, ciphertext has changed, this could indicate incorrect key")
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return fmt.Errorf("Error creating AES block: %v", err)
	}

	iv := e.CipherText[:aes.BlockSize]
	e.CipherText = e.CipherText[aes.BlockSize:]

	e.PlainText = make([]byte, len(e.CipherText))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(e.PlainText, e.CipherText)

	return nil
}

// CreateHMAC will generate a cryptographic hash of data supplied
func CreateHMAC(keyb []byte, data []byte) ([]byte, error) {
	h := hmac.New(sha512.New, keyb)
	_, err := h.Write(data)
	if err != nil {
		return nil, fmt.Errorf("Error writing data for HMAC: %v", err)
	}

	return h.Sum(nil), nil
}

// WrapCrypto wraps both cipher text and hmac into a single string
// to be written to disk
func (e *EncryptionObject) WrapCrypto() {
	b64cipher := base64.StdEncoding.EncodeToString(e.CipherText)
	b64hmac := base64.StdEncoding.EncodeToString(e.HMAC)

	e.WrappedData = fmt.Sprintf("@encrypted_data(%s)\n@hmac(%s)", b64cipher, b64hmac)
}

//UnwrapCrypto unwraps cipher text and hmac to allow decryption
func (e *EncryptionObject) UnwrapCrypto() error {
	var err error

	b64cipher := wrappedCipherRegex.FindStringSubmatch(e.WrappedData)
	if b64cipher == nil || len(b64cipher) < 1 || b64cipher[1] == "" {
		return fmt.Errorf("unwrapping cipher text")
	}

	e.CipherText, err = base64.StdEncoding.DecodeString(b64cipher[1])
	if err != nil {
		return fmt.Errorf("decoding base64 cipher: %v", err)
	}

	b64hmac := wrappedHmacRegex.FindStringSubmatch(e.WrappedData)
	if b64hmac == nil || len(b64hmac) < 1 || b64hmac[1] == "" {
		return fmt.Errorf("unwrapping HMAC text")
	}

	e.HMAC, err = base64.StdEncoding.DecodeString(b64hmac[1])
	if err != nil {
		return fmt.Errorf("decoding base64 HMAC: %v", err)
	}

	return nil
}

// RandomKey returns a random password of specified length
func RandomKey(n int) string {
	const chars = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`
	b := make([]byte, n)

	rdm := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = chars[rdm.Intn(len(chars))]
	}

	return string(b)
}

func RandomKeyB64(n int) string {
	b := make([]byte, n)
	rand.Reader.Read(b)

	return base64.StdEncoding.EncodeToString(b)
}

func (e *EncryptionObject) ReadConfigFiles(filename string) []byte {
	var (
		file []byte
		err  error
	)
	if filename != "" {
		file, err = ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal("Error reading file: ", err)
		}
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting working directory: %v", err)
		}
		files, err := ioutil.ReadDir(pwd)
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".vc") {
				fbytes, err := ioutil.ReadFile(f.Name())
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
				file = JoinBytes(fbytes, file)
			}
		}
	}
	return file
}

func (e *EncryptionObject) ReadEncryptedConfigFiles(filename string) []byte {
	var (
		file []byte
		err  error
	)
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}
	files, err := ioutil.ReadDir(pwd)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".vc.enc") {
			fbytes, err := ioutil.ReadFile(f.Name())
			if err != nil {
				log.Fatalf("Error reading file: %v", err)
			}

			e.WrappedData = string(fbytes)
			if err := e.UnwrapCrypto(); err != nil {
				log.Fatalf("Error unwrapping encrypted file: %v\nErr: %v", f.Name(), err)
			}
			if err := e.Decrypt(); err != nil {
				log.Fatalf("Error decrypting file: %v\n Err: %v", f.Name(), err)
			}
			file = JoinBytes([]byte(e.PlainText), file)
		}
	}
	return file
}

func EncryptString(data string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Error creating AES block: %v", err)
	}

	pt := []byte(data)
	ct := make([]byte, aes.BlockSize+len(data))
	iv := ct[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("Error creating iv: %v", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ct[aes.BlockSize:], pt)

	return base64.StdEncoding.EncodeToString(ct), nil
}

func DecryptString(cipherText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Error creating AES block: %v", err)
	}

	ct, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("Error base64 decoding cipher text: %v", err)
	}
	iv := ct[:aes.BlockSize]
	ct = ct[aes.BlockSize:]

	pt := make([]byte, len(ct))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(pt, ct)

	return string(pt), nil
}

func JoinBytes(dst, src []byte) []byte {
	for _, b := range src {
		dst = append(dst, b)
	}
	dst = append(dst, byte(10))

	return dst
}

func GetPassword() ([]byte, error) {
	fmt.Print("Please enter key: ")
	bytesB64Key, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("Error reading encryption key from terminal: %v", err)
	}
	bytesKey, err := base64.StdEncoding.DecodeString(string(bytesB64Key))
	if err != nil {
		return nil, fmt.Errorf("Error decoding base64 key: %v", err)
	}

	return bytesKey, nil
}
