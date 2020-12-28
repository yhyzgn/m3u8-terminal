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
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
)

type Downloader struct {
	wg         *sync.WaitGroup
	pool       chan *Resource
	Concurrent int
	Client     *http.Client
	Dir        string
	Resources  []*Resource
}

func New(dir string) *Downloader {
	_ = os.MkdirAll(dir, os.ModePerm)
	concurrent := runtime.NumCPU()

	return &Downloader{
		wg:         &sync.WaitGroup{},
		pool:       make(chan *Resource, concurrent),
		Concurrent: concurrent,
		Client:     http.DefaultClient,
		Dir:        dir,
	}
}

func (dl *Downloader) AppendResource(url, filename string) *Downloader {
	dl.Resources = append(dl.Resources, NewResource(url, filename))
	return dl
}

func (dl *Downloader) Append(resources ...*Resource) *Downloader {
	dl.Resources = append(dl.Resources, resources...)
	return dl
}

func (dl *Downloader) StartWithReader(reader func(resourceIndex int, reader io.ReadCloser) io.Reader) {
	fmt.Println("Downloader started, concurrent is ", dl.Concurrent)

	p := mpb.New(mpb.WithWaitGroup(dl.wg))

	for i, task := range dl.Resources {
		dl.wg.Add(1)
		task.index = i
		go dl.download(task, p, reader)
	}
	p.Wait()
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
	finalPath := path.Join(dl.Dir, resource.Filename)
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
	resp, err := dl.Client.Do(req)
	if nil != err {
		return
	}

	defer resp.Body.Close()

	// 获取到文件大小
	fileSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

	// 创建一个进度条
	bar := progress.AddBar(fileSize,
		mpb.PrependDecorators(
			decor.Name(resource.Filename, decor.WC{W: len(resource.Filename) + 1, C: decor.DidentRight}),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		),
	)

	proxyReader := bar.ProxyReader(resp.Body)
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
	<-dl.pool
	return
}
