// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-28 11:48
// version: 1.0.0
// desc   :
//
// http://devimages.apple.com/iphone/samples/bipbop/bipbopall.m3u8
// http://devimages.apple.com/iphone/samples/bipbop/gear1/prog_index.m3u8
// http://devimages.apple.com/iphone/samples/bipbop/gear1/fileSequence0.ts

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"m3u8/built"
	"m3u8/crypt"
	"m3u8/settings"
	"path"
)

var (
	conf           settings.Config
	supportedMedia = map[string]bool{
		"mp4": true,
		"mkv": true,
		"avi": true,
	}
	url  = flag.String("url", "", "URL of m3u8 resource")
	name = flag.String("name", "", "Name of media")
	ext  = flag.String("ext", "", "Extension of media")
)

func init() {
	fmt.Println(built.FullName)

	bs, err := ioutil.ReadFile("./settings.json")
	if nil != err {
		panic(err)
	}
	if err = json.Unmarshal(bs, &conf); nil != err {
		panic(err)
	}
}

func main() {
	flag.Parse()

	// 控制台参数
	urlStr := *url
	filename := *name
	fileExt := *ext

	if "" == urlStr {
		panic("URL of m3u8 resource can not be empty.")
	}

	md5Filename := crypt.MD5String(urlStr)
	if "" == filename {
		filename = md5Filename
	}

	if "" == fileExt {
		// 默认扩展
		fileExt = conf.Extension
	}
	if !supportedMedia[fileExt] {
		panic("Unsupported extension of media: " + fileExt)
	}

	// 检查 ffmpeg
	checkFfmpeg()

	saveDir := conf.SaveDir
	tsDir := path.Join(saveDir, conf.TsTempDirPrefix+md5Filename)
	mediaFile := filename + "." + fileExt

	mediaPath := path.Join(saveDir, mediaFile)

	if should := shouldDownload(mediaPath); !should {
		fmt.Println(fmt.Sprintf("Mabey the media %s Exists", colorful(mediaFile)))
		return
	}

	// 下载任务，返回已下载成功的切片列表
	tsFile := download(urlStr, tsDir, mediaFile)

	// 下载完成，开始合并
	fmt.Println("TS files download finished, now merging...")

	// 合并切片，并转换视频格式
	merge(tsDir, mediaPath, mediaFile, tsFile)
}
