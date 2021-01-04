package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"sort"
	"strings"

	"github.com/fbsobreira/gotron-sdk/pkg/common"
)

const (
	mark4 int64 = (1 << 4) - 1
)

func Sha1Hash(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte{}))
}

func Sha1Sign(strs ...string) string {
	sort.Strings(strs)
	return Sha1Hash(strings.Join(strs, ""))
}

func Sha256Hash(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte{}))
}

func Sha256Sign(strs ...string) string {
	sort.Strings(strs)
	return Sha256Hash(strings.Join(strs, ""))
}

func GetEntityType(id int64) int32 {
	return int32((id >> 12) & mark4)
}

func GlcEncode(id int64) string {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(id))
	return common.EncodeCheck(b)
}

func GlcDecode(str string) (int64, error) {
	b, err := common.DecodeCheck(str)
	if err != nil {
		return 0, err
	}

	return int64(binary.LittleEndian.Uint64(b)), nil
}

func EncodeAddress(id int64, chainId int32) string {
	b := make([]byte, 20)
	binary.LittleEndian.PutUint64(b, uint64(id))
	binary.LittleEndian.PutUint64(b[8:], ^uint64(id))
	binary.LittleEndian.PutUint32(b[16:], uint32(chainId))
	return common.EncodeCheck(b)
}

func DecodeAddress(address string) (int64, int32, error) {
	b, err := common.DecodeCheck(address)
	if err != nil {
		return 0, 0, err
	}

	id := int64(binary.LittleEndian.Uint64(b))
	idr := binary.LittleEndian.Uint64(b[8:])
	if ^uint64(id) != idr {
		return 0, 0, errors.New("address invalid")
	}

	chainId := int32(binary.LittleEndian.Uint32(b[16:]))
	return id, chainId, nil
}

func GetOpenId(userId int64, appId string) (string, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(userId))

	crypted, err := AesEncrypt(b, []byte(appId))
	if err != nil {
		return "", nil
	}
	return common.EncodeCheck(crypted), nil
}

func GetUserId(openId string, appId string) (int64, error) {
	crypted, err := common.DecodeCheck(openId)
	if err != nil {
		return 0, err
	}

	b, err := AesDecrypt(crypted, []byte(appId))
	if err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(b)), nil
}

func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func keyHash(key []byte) []byte {
	c := sha256.New()
	c.Write(key)
	return c.Sum(nil)
}

func AesEncrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyHash(key))
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	data = pkcs7Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyHash(key))
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs7UnPadding(origData)
	return origData, nil
}
