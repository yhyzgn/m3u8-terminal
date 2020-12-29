// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-28 16:44
// version: 1.0.0
// desc   : 

package dl

import (
	"fmt"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"github.com/yhyzgn/golus"
	"io"
	"m3u8/file"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
)

type Downloader struct {
	wg           *sync.WaitGroup
	pool         chan *Resource
	concurrent   int
	client       *http.Client
	dir          string
	resources    []*Resource
	showProgress bool
	finished     chan *Resource
}

func New(dir string) *Downloader {
	if err := os.MkdirAll(dir, os.ModePerm); nil != err {
		panic(err)
	}

	concurrent := runtime.NumCPU()

	return &Downloader{
		wg:           &sync.WaitGroup{},
		pool:         make(chan *Resource, concurrent),
		concurrent:   concurrent,
		client:       http.DefaultClient,
		dir:          dir,
		showProgress: true,
		finished:     make(chan *Resource, concurrent),
	}
}

func (dl *Downloader) AppendResource(url, filename string) *Downloader {
	dl.resources = append(dl.resources, NewResource(url, filename, true))
	return dl
}

func (dl *Downloader) Append(resources ...*Resource) *Downloader {
	dl.resources = append(dl.resources, resources...)
	return dl
}

func (dl *Downloader) ShowProgressBar(show bool) *Downloader {
	dl.showProgress = show
	return dl
}

func (dl *Downloader) Finished() chan *Resource {
	return dl.finished
}

func (dl *Downloader) StartWithReader(reader func(resourceIndex int, reader io.ReadCloser) io.Reader) {
	fmt.Println("Downloader started, concurrent is ", dl.concurrent)
	var progress *mpb.Progress
	if dl.showProgress {
		progress = mpb.New(mpb.WithWaitGroup(dl.wg))
	}
	for i, task := range dl.resources {
		dl.wg.Add(1)
		task.index = i
		go dl.download(task, progress, reader)
	}
	if nil != progress {
		progress.Wait()
	}
	dl.wg.Wait()
}

func (dl *Downloader) Start() {
	dl.StartWithReader(func(resourceIndex int, reader io.ReadCloser) io.Reader {
		return reader
	})
}

func (dl *Downloader) download(resource *Resource, progress *mpb.Progress, reader func(resourceIndex int, reader io.ReadCloser) io.Reader) (err error) {
	defer dl.wg.Done()
	dl.pool <- resource
	finalPath := path.Join(dl.dir, resource.Filename)

	// 如果不覆盖下载，文件存在时则无需下载
	if file.Exists(finalPath) && !resource.Overwrite {
		// 也表示完成一个任务
		dl.finished <- <-dl.pool
		return
	}

	tempPath := finalPath + ".tmp"

	// 创建一个临时文件
	target, err := os.Create(tempPath)
	if nil != err {
		return
	}

	// 开始下载
	req, err := http.NewRequest(http.MethodGet, resource.URL, nil)
	if nil != err {
		return
	}
	resp, err := dl.client.Do(req)
	if nil != err {
		return
	}

	defer resp.Body.Close()
	proxyReader := resp.Body
	// 创建一个进度条
	if nil != progress {
		// 获取到文件大小
		fileSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		bar := progress.AddBar(fileSize,
			mpb.BarStyle("[=>_]<+"),
			mpb.BarFillerClearOnComplete(),
			mpb.PrependDecorators(
				decor.Name(resource.Filename, decor.WC{W: len(resource.Filename) + 1, C: decor.DidentRight}),
				decor.CountersKibiByte("% .2f / % .2f", decor.WC{W: 32}),
				decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth), ""),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.NewPercentage("%.2f", decor.WC{W: 7}), "  "+golus.NewStylus().SetFontColor(golus.FontGreen).SetFontStyle(golus.StyleBold).Apply("Download Finished")),
			),
		)
		proxyReader = bar.ProxyReader(proxyReader)
	}

	defer proxyReader.Close()
	realReader := reader(resource.index, proxyReader)

	// 将下载的文件流写到临时文件
	_, err = io.Copy(target, realReader)
	if nil != err {
		_ = target.Close()
		return
	}

	_ = target.Close()
	err = os.Rename(tempPath, finalPath)
	if nil != err {
		return
	}

	// 完成一个任务
	dl.finished <- <-dl.pool
	return
}
