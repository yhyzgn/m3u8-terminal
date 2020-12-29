// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-29 11:24
// version: 1.0.0
// desc   : 

package file

import "os"

// 检测文件路径是否存在
//
// 如果返回的错误为nil,说明文件或文件夹存在
// 如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
// 如果返回的错误为其它类型,则不确定是否在存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}