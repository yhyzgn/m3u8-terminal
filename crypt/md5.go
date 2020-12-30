// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-30 9:14
// version: 1.0.0
// desc   : 

package crypt

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5String 获取字符串 md5 值
func MD5String(s string) string {
	return MD5Byte([]byte(s))
}

// MD5Byte 获取字节数组 md5 值
func MD5Byte(s []byte) string {
	h := md5.New()
	h.Write(s)
	return hex.EncodeToString(h.Sum(nil))
}
