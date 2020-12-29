// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-29 9:26
// version: 1.0.0
// desc   : 

package dl

import (
	"fmt"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"math/rand"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// initialize progress container, with custom width
	p := mpb.New(mpb.WithWidth(64))

	total := 100
	name := "Single Bar:"
	// adding a single bar, which will inherit container's width
	bar := p.AddBar(int64(total),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), "done",
			),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	// simulating some work
	max := 100 * time.Millisecond
	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(rand.Intn(10)+1) * max / 10)
		bar.Increment()
	}
	// wait for our bar to complete and flush
	p.Wait()
}

func TestDownloader_Start(t *testing.T) {
	downloader := New("./tst").ShowProgressBar(false)
	p := mpb.New(mpb.WithWidth(64))

	total := 3
	name := "测试："
	bar := p.AddBar(int64(total),
		// override DefaultBarStyle, which is "[=>-]<+"
		// mpb.BarStyle("╢▌▌░╟"),
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.NewPercentage("%.2f", decor.WC{W: 7}), "  \x1b[32;1;4mdone\x1b[0m"),
		),
	)

	for i := 0; i < total; i++ {
		downloader.AppendResource("https://ime.sogoucdn.com/7ef36e63db762cb1ec1bd8575bda2c6c/5fea87fe/dl/index/1603177583/sogou_pinyin_98a.exe", fmt.Sprintf("搜狗_%.6d.exe", i+1))
	}

	go func() {
		for {
			<-downloader.Finished()
			bar.Increment()
		}
	}()
	go downloader.Start()
	p.Wait()
}
