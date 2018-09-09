package method

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/moxianfeng/gofactory"
)

type aes_256_cfb_encrypt struct {
	MethodBase
}

func (this *aes_256_cfb_encrypt) Process(in []byte) ([]byte, error) {
	if len(this.Arguments) < 32 {
		return nil, aes.KeySizeError(len(this.Arguments))
	}

	key := []byte(this.Arguments[:32])
	block, err := aes.NewCipher(key)
	if nil != err {
		return nil, err
	}
	paddingOrig := PKCS7Padding(in, block.BlockSize())

	cipherText := make([]byte, block.BlockSize()+len(paddingOrig))
	iv := cipherText[:block.BlockSize()]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCFBEncrypter(block, iv)
	mode.XORKeyStream(cipherText[block.BlockSize():], paddingOrig)

	// log.Printf("encrypt(%v) = %v", in, cipherText)
	return cipherText, nil
}

type aes_256_cfb_decrypt struct {
	MethodBase
}

func (this *aes_256_cfb_decrypt) Process(in []byte) ([]byte, error) {
	// args is key
	if len(this.Arguments) < 32 {
		return nil, aes.KeySizeError(len(this.Arguments))
	}
	key := []byte(this.Arguments[:32])
	block, err := aes.NewCipher(key)
	if nil != err {
		return nil, err
	}
	if len(in) < block.BlockSize() {
		return nil, fmt.Errorf("Data too short for aes(%d), in:%v", len(in), in)
	}

	iv := in[:block.BlockSize()]
	mode := cipher.NewCFBDecrypter(block, iv)

	orig := make([]byte, len(in[block.BlockSize():]))
	mode.XORKeyStream(orig, in[block.BlockSize():])

	ret, err := PKCS7UnPadding(orig)
	// log.Printf("decrypt(%v) = %v", in, ret)
	return ret, err
}

func init() {
	gofactory.Default.Register("aes_256_cfb_encrypt", &aes_256_cfb_encrypt{})
	gofactory.Default.Register("aes_256_cfb_decrypt", &aes_256_cfb_decrypt{})
}
