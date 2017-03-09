package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/yahoojapan/yisucon/benchmarker/logger"
)

func GetMD5(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetMD5ByIO(r io.Reader) string {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		logger.GetLogger().Println(err)
	}
	return GetMD5(bytes)
}

func UIDGen() (string, error) {
	buf := make([]byte, 10)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d%x", time.Now().UnixNano(), buf[0:len(buf)]))), nil
}

func Cipher(text string) string {
	shift, offset := rune(1), rune(26)

	runes := []rune(text)
	for index, char := range runes {
		if char >= 'a'-shift && char <= 'z'-shift ||
			char >= 'A'-shift && char <= 'Z'-shift {
			char += shift
		} else {
			char = char + shift - offset
		}
		runes[index] = char
	}
	return string(runes)
}
