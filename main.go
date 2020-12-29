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
	"fmt"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"io"
	"io/ioutil"
	"m3u8/crypt"
	"m3u8/dl"
	"m3u8/file"
	"m3u8/http"
	"m3u8/list"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func main() {
	urlStr := "http://devimages.apple.com/iphone/samples/bipbop/gear1/prog_index.m3u8"
	saveDir := "./down"
	filename := "测试"
	tsDir := path.Join(saveDir, "ts_"+filename)
	mediaFile := path.Join(saveDir, filename+".mp4")

	_, mediaList, err := list.GetPlayList(urlStr)
	if nil != err {
		fmt.Println(err)
		return
	}
	if nil == mediaList {
		fmt.Println("No any media source found.")
	}

	progress := mpb.New(
		mpb.WithWidth(100),
		mpb.WithRefreshRate(time.Second),
	)
	bar := progress.AddBar(int64(len(mediaList.Segments)),
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(
			decor.Name(filename, decor.WC{W: len(filename) + 1, C: decor.DidentRight}),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.NewPercentage("%.2f", decor.WC{W: 7}), "  "+colorful("Download Finished")),
		),
	)

	downloader := dl.New(tsDir).ShowProgressBar(false)

	keyMap := make(map[string][]byte)
	tsNames := make([]string, 0)
	for i, seg := range mediaList.Segments {
		if nil != seg {
			if nil != seg.Key && seg.Key.URI != "" && nil == keyMap[seg.Key.Method+"-"+seg.Key.URI] {
				keyMap[seg.Key.Method+"-"+seg.Key.URI], _ = http.Get(seg.Key.URI)
			}
			name := fmt.Sprintf("slice_%.6d.ts", i+1)
			tsNames = append(tsNames, path.Join(tsDir, name))
			downloader.AppendResource(seg.URI, name)
		}
	}

	// 更新进度条
	go func() {
		for {
			<-downloader.Finished()
			bar.Increment()
		}
	}()

	go downloader.StartWithReader(func(resourceIndex int, reader io.ReadCloser) io.Reader {
		key := mediaList.Segments[resourceIndex].Key
		if nil == key {
			return reader
		}
		data, _ := ioutil.ReadAll(reader)
		data, _ = crypt.AES128Decrypt(data, keyMap[key.Method+"-"+key.URI], []byte(key.IV))
		return bytes.NewReader(data)
	})
	// 等待任务完成
	progress.Wait()

	// 下载完成，开始合并
	fmt.Println("TS files download finished, now merging...")

	if file.Exists(mediaFile) {
		var ch string
		fmt.Print("Media file exist, cover? (y/n) ")
		_, err = fmt.Scan(&ch)
		if nil != err {
			fmt.Println(err)
			return
		}

		if ch == "n" {
			err = os.RemoveAll(tsDir)
			if nil != err {
				fmt.Println(err)
			} else {
				fmt.Println(fmt.Sprintf("%s Media Exist", colorful(filename)))
			}
			return
		}

		// 删除已存在文件
		err = os.Remove(mediaFile)
		if nil != err {
			fmt.Println(err)
			return
		}
		fmt.Println("Merging...")
	}

	// ffmpeg -i "concat:file001.ts|file002.ts|file003.ts|file004.ts......n.ts" -acodec copy -vcodec copy -absf aac_adtstoasc out.mp4
	concat := "concat:" + strings.Join(tsNames, "|")
	cmdArgs := []string{"-i", concat, "-acodec", "copy", "-vcodec", "copy", "-absf", "aac_adtstoasc", mediaFile}

	cmd := exec.Command("ffmpeg", cmdArgs...)
	cmd.Stdout = os.Stdout
	if err = cmd.Run(); nil == err {
		// 合并完成，删除ts目录
		err = os.RemoveAll(tsDir)
		if nil != err {
			fmt.Println(err)
		} else {
			fmt.Println(fmt.Sprintf("%s Media %s", colorful(filename), colorful("Merge Finished")))
		}
	}
}

func colorful(msg string) string {
	return fmt.Sprintf("\x1b[32;1;4m%s\x1b[0m", msg)
}
