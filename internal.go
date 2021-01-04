// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-29 16:14
// version: 1.0.0
// desc   : 

package main

import (
	"bytes"
	"fmt"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"github.com/yhyzgn/golus"
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
	"runtime"
	"strings"
	"unicode/utf8"
)

const (
	ffmpeg    = "ffmpeg" // ffmpeg 命令
	ffmpegDir = "./"     // ffmpeg 可执行程序所在目录，为达到环境变量优先，这里设置为当前程序运行的目录
)

// 检查 ffmpeg
func checkFfmpeg() {
	// 预先检查程序是否存在
	if _, err := exec.LookPath(ffmpeg); nil != err {
		// 下载 ffmpeg
		gs := runtime.GOOS
		switch gs {
		case "windows":
			ffmpegDownload(conf.Ffmpeg.Windows, ffmpeg+".exe")
		case "darwin":
			ffmpegDownload(conf.Ffmpeg.Mac, ffmpeg)
			if err = exec.Command("chmod", "+X", ffmpeg).Run(); nil != err {
				panic(err)
			}
		case "linux":
			ffmpegDownload(conf.Ffmpeg.Linux, ffmpeg)
			if err = exec.Command("chmod", "+X", ffmpeg).Run(); nil != err {
				panic(err)
			}
		default:
			fmt.Println("Unknown os: ", gs)
		}
	}
}

// 下载 ffmpeg
func ffmpegDownload(url, name string) {
	dl.New(ffmpegDir).AppendResource(url, name).Start()
}

// 检查文件是否存在，并根据控制台输入参数决定是否应该覆盖下载
func shouldDownload(mediaPath string) (should bool) {
	if file.Exists(mediaPath) {
		var ch string
		fmt.Print(golus.NewStylus().SetFontColor(golus.FontYellow).Apply(fmt.Sprintf("Media file '%s' already exists, Overwrite? (y/n) ", mediaPath)))
		_, err := fmt.Scan(&ch)
		if nil != err {
			ch = "n"
		}

		// 不覆盖
		if ch == "n" {
			should = false
			return
		}

		// 删除已存在文件
		err = os.Remove(mediaPath)
		if nil != err {
			should = false
			return
		}
	}
	should = true
	return
}

// 下载 m3u8 资源
func download(urlStr, tsDir, mediaFile string) (tsNames []string, tsFile string) {
	_, mediaList, err := list.GetPlayList(urlStr)
	if nil != err {
		fmt.Println(err)
		return
	}
	if nil == mediaList {
		fmt.Println("No any media source found.")
	}

	progress := mpb.New(
		mpb.WithWidth(160),
	)
	bar := progress.AddBar(int64(len(mediaList.Segments)),
		mpb.BarStyle("[=>_]<+"),
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(
			decor.Name("Download -- "),
			decor.Name(colorful(mediaFile), decor.WC{W: utf8.RuneCountInString(mediaFile), C: decor.DidentRight}),
			decor.CountersNoUnit(" %06d / %06d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.NewPercentage("%.2f", decor.WC{W: 7}), colorful("Download Finished")),
		),
	)

	downloader := dl.New(tsDir).ShowProgressBar(false)

	keyMap := make(map[string][]byte)
	tsNames = make([]string, 0)
	for i, seg := range mediaList.Segments {
		if nil != seg {
			if nil != seg.Key && seg.Key.URI != "" && nil == keyMap[seg.Key.Method+"-"+seg.Key.URI] {
				keyMap[seg.Key.Method+"-"+seg.Key.URI], _ = http.Get(seg.Key.URI)
			}
			name := fmt.Sprintf("slice_%06d.ts", i+1)
			tsNames = append(tsNames, "file "+name)
			downloader.Append(dl.NewResource(seg.URI, name, false))
		}
	}

	tsFile = path.Join(tsDir, "slice.lst")
	_ = file.WriteString(tsFile, strings.Join(tsNames, "\n"))

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
	return
}

// 合并切片，并转换视频格式
func mergeByFfmpeg(tsDir, mediaPath, mediaFile, tsFile string) {
	// ffmpeg -i "xxx.txt" -acodec copy -vcodec copy -absf aac_adtstoasc out.mp4
	cmdArgs := []string{"-y", "-f", "concat", "-i", tsFile, "-acodec", "copy", "-vcodec", "copy", "-absf", "aac_adtstoasc", mediaPath}

	cmd := exec.Command(ffmpeg, cmdArgs...)

	if err := cmd.Run(); err == nil {
		// 合并完成，删除ts目录
		err := os.RemoveAll(tsDir)
		if nil != err {
			fmt.Println(err)
		} else {
			fmt.Println(fmt.Sprintf("Media '%s' Merge Finished", colorful(mediaFile)))
		}
	} else {
		fmt.Println(err)
	}
}

// 直接合并成ts文件
func merge(tsDir, mediaPath, mediaFile string, tsNames []string) {
	if nil == tsNames {
		return
	}

	var err error
	for _, name := range tsNames {
		tsFile := path.Join(tsDir, name[5:])
		bs, e := file.Read(tsFile)
		if nil != e {
			err = e
			break
		}
		e = file.Append(mediaPath, bs)
		if nil != e {
			err = e
			break
		}
	}
	if nil == err {
		// 合并完成，删除ts目录
		err := os.RemoveAll(tsDir)
		if nil != err {
			fmt.Println(err)
		} else {
			fmt.Println(fmt.Sprintf("Media '%s' Merge Finished", colorful(mediaFile)))
		}
	} else {
		fmt.Println(err)
	}
}

// 彩色控制台输出
func colorful(msg string) string {
	return golus.NewStylus().SetFontColor(golus.FontGreen).SetFontStyle(golus.StyleBold, golus.StyleUnderLine).Apply(msg)
}
