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
	"time"
)

const (
	ffmpeg    = "ffmpeg" // ffmpeg 命令
	ffmpegDir = "./"     // ffmpeg 可执行程序所在目录，为达到环境变量优先，这里设置为当前程序运行的目录
)

// 检查 ffmpeg
func checkFfmpeg() {
	// 先运行一波，优先使用环境变量中的配置
	// 如果无法识别则重新下载到指定目录
	if err := exec.Command(ffmpeg).Run(); nil != err && strings.Contains(err.Error(), "not found") {
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
		fmt.Print(golus.NewStylus().SetFontColor(golus.FontYellow).Apply("Media file exist, cover? (y/n) "))
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
func download(urlStr, tsDir, mediaFile string) (tsNames []string) {
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
			decor.Name("Download -- "),
			decor.Name(mediaFile, decor.WC{W: len(mediaFile) + 1, C: decor.DidentRight}),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.NewPercentage("%.2f", decor.WC{W: 7}), "  "+colorful("Download Finished")),
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
	return
}

// 合并切片，并转换视频格式
func merge(tsDir, mediaPath, mediaFile string, tsNames []string) {
	// ffmpeg -i "concat:file001.ts|file002.ts|file003.ts|file004.ts......n.ts" -acodec copy -vcodec copy -absf aac_adtstoasc out.mp4
	concat := "concat:" + strings.Join(tsNames, "|")
	cmdArgs := []string{"-i", concat, "-acodec", "copy", "-vcodec", "copy", "-absf", "aac_adtstoasc", mediaPath}

	cmd := exec.Command(ffmpeg, cmdArgs...)
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); nil == err {
		// 合并完成，删除ts目录
		err = os.RemoveAll(tsDir)
		if nil != err {
			fmt.Println(err)
		} else {
			fmt.Println(fmt.Sprintf("%s Media %s", colorful(mediaFile), colorful("Merge Finished")))
		}
	}
}

// 彩色控制台输出
func colorful(msg string) string {
	return golus.NewStylus().SetFontColor(golus.FontGreen).SetFontStyle(golus.StyleBold, golus.StyleUnderLine).Apply(msg)
}
