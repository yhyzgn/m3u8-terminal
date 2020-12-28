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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"m3u8/crypt"
	"m3u8/dl"
	"m3u8/http"
	"m3u8/list"
	"m3u8/m3u8"
	"os"
	"os/exec"
	"path"
	"strings"
)

func main() {
	urlStr := "http://devimages.apple.com/iphone/samples/bipbop/gear1/prog_index.m3u8"
	saveDir := "./down"
	filename := "测试"
	tsDir := path.Join(saveDir, "ts_"+filename)

	masterList, mediaList, err := list.GetPlayList(urlStr)
	if nil != err {
		fmt.Println(err)
		return
	}

	if nil != masterList {
		// master
		vnt, err := chooseStream(masterList)
		if nil != err {
			fmt.Println(err)
			return
		}
		masterList, mediaList, err = list.GetPlayList(vnt.URI)
		if nil != err {
			fmt.Println(err)
			return
		}
	}

	downloader := dl.New(tsDir)

	keyMap := make(map[string][]byte)
	tsNames := make([]string, 0)

	for i, v := range mediaList.Segments {
		if v != nil {
			if nil != v.Key && v.Key.URI != "" && nil == keyMap[v.Key.Method+"-"+v.Key.URI] {
				keyMap[v.Key.Method+"-"+v.Key.URI], _ = http.Get(v.Key.URI)
			}
			tsName := fmt.Sprintf("file_%d.ts", i)
			tsNames = append(tsNames, path.Join(tsDir, tsName))
			downloader.AppendResource(v.URI, tsName)
		}
	}

	downloader.StartWithReader(func(resourceIndex int, reader io.ReadCloser) io.Reader {
		key := mediaList.Segments[resourceIndex].Key
		if nil == key {
			return reader
		}
		data, _ := ioutil.ReadAll(reader)
		data, _ = crypt.AES128Decrypt(data, keyMap[key.Method+"-"+key.URI], []byte(key.IV))
		return bytes.NewReader(data)
	})

	// ffmpeg -i "concat:file001.ts|file002.ts|file003.ts|file004.ts......n.ts" -acodec copy -vcodec copy -absf aac_adtstoasc out.mp4
	concat := "concat:" + strings.Join(tsNames, "|")
	cmdArgs := []string{"-i", concat, "-acodec", "copy", "-vcodec", "copy", "-absf", "aac_adtstoasc", path.Join(saveDir, filename+".mp4")}

	cmd := exec.Command("./script/ffmpeg", cmdArgs...)
	if err = cmd.Run(); nil == err {
		// 合并完成，删除ts目录
		err = os.RemoveAll(tsDir)
		if nil != err {
			fmt.Println(err)
		} else {
			fmt.Println(fmt.Sprintf("【%s】下载完成", filename))
		}
	}
}

func chooseStream(mediaList *m3u8.MasterPlaylist) (vnt *m3u8.Variant, err error) {
	var sb strings.Builder
	sb.WriteString("Please choose program: \n")
	for i, vnt := range mediaList.Variants {
		sb.WriteString(fmt.Sprintf("\t%d. BandWidth: %d, Resolution: %s\n", i+1, vnt.Bandwidth, vnt.Resolution))
	}
	sb.WriteString("Input the No. you want: ")
	var index int
	fmt.Print(sb.String())
	_, err = fmt.Scan(&index)
	if nil != err {
		return
	}
	if index <= 0 || index > len(mediaList.Variants) {
		err = errors.New("input index out of variants range")
		return
	}
	vnt = mediaList.Variants[index-1]
	fmt.Println(fmt.Sprintf("You choice is: BandWidth: %d, Resolution: %s", vnt.Bandwidth, vnt.Resolution))
	return
}
