package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"github.com/pkg/errors"
	"github.com/spaolacci/murmur3"
	"golang.org/x/crypto/bcrypt"
)

func ValidatePassword(userPassWord string, hashed string) (bool, error) { // 解密
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(userPassWord)); err != nil {
		return false, errors.New("ValidatePassword: passWord is not correct")
	}
	return true, nil
}

func GeneratePassWord(userPassword string) ([]byte, error) { // 加密
	return bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
}

// 高级加密标准 （Advanced EnCryption Standard， AES）
// 16,24,32位字符串，分别对应AES-128，AES-192，AES-256加密方法
// key不能泄漏
var PwdKey = []byte("DIS**#KKKDJJSKDI")

// PKCS7 填充模式
func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	// bytes.Repeat()函数的功能是把切片[]byte{byte(padding)}复制padding个，然后合并成新的字节返回
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// 填充的反向操作
func pkcs7UnPadding(origData []byte) ([]byte, error) {
	//获取数据长度
	length := len(origData)
	if length == 0 {
		return nil, errors.New("加密字符串错误")
	} else {
		// 获取填充字符串长度
		unPadding := int(origData[length-1])
		//截取切片，删除填充字节，并且返回明文
		return origData[:(length - unPadding)], nil
	}
}

func aesEncrypt(origData []byte, key []byte) ([]byte, error) {
	// 1.创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "encryption#AesEncrypt: encrypt failed")
	}

	// 2.获取块的大小
	blockSize := block.BlockSize()

	// 3.对加密数据进行填充，让加密数据满足需求
	origData = pkcs7Padding(origData, blockSize)

	// 4.采用AES加密方法中的CBC加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))

	// 5.执行加密
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 实现解密
func aesDeCrypt(cypted []byte, key []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块大小
	blockSize := block.BlockSize()
	//创建加密客户端实例
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(cypted))
	//这个函数也可以用来解密
	blockMode.CryptBlocks(origData, cypted)
	//去除填充字符串
	origData, err = pkcs7UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, err
}

// 加密base64
func EnPwdCode(pwd []byte) (string, error) {
	result, err := aesEncrypt(pwd, PwdKey)
	if err != nil {
		return "", errors.Wrap(err, "encryption#EnPwdCode: encrypt failed")
	}

	return base64.StdEncoding.EncodeToString(result), nil
}

//解密
func DePwdCode(pwd string) ([]byte, error) {
	//解密base64字符串
	pwdByte, err := base64.StdEncoding.DecodeString(pwd)
	if err != nil {
		return nil, errors.Wrap(err, "encryption#DePwdCode: decrypt failed")
	}
	//执行AES解密
	return aesDeCrypt(pwdByte, PwdKey)
}

// Hash returns the hash value of data.
func Hash(data []byte) uint64 {
	return murmur3.Sum64(data)
}
