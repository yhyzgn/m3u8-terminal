// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-28 17:51
// version: 1.0.0
// desc   :

package list

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/grafov/m3u8"
	"github.com/yhyzgn/golus"
	"m3u8/http"
	"net/url"
	"strings"
)

func GetPlayList(urlStr string) (masterList *m3u8.MasterPlaylist, mediaList *m3u8.MediaPlaylist, err error) {
	_, err = url.Parse(urlStr)
	if nil != err {
		return
	}

	bs, err := http.Get(urlStr)
	if nil != err {
		return
	}

	playList, listType, err := m3u8.Decode(*bytes.NewBuffer(bs), true)
	if nil != err {
		return
	}

	if listType == m3u8.MEDIA {
		// media
		mediaList = playList.(*m3u8.MediaPlaylist)
		// mediaList 中存在 nil 的 Segments，需要筛选出有效的 Segments
		segList := make([]*m3u8.MediaSegment, 0)

		var lastKey *m3u8.Key
		for _, seg := range mediaList.Segments {
			if nil != seg {
				seg.URI = completeURI(urlStr, seg.URI)
				if nil == seg.Key && nil != lastKey {
					seg.Key = lastKey
				}
				if nil != seg.Key && seg.Key != lastKey {
					lastKey = seg.Key
					lastKey.URI = completeURI(urlStr, lastKey.URI)
				}
				segList = append(segList, seg)
			}
		}
		mediaList.Segments = segList
	} else {
		// master
		masterList = playList.(*m3u8.MasterPlaylist)
		for _, vnt := range masterList.Variants {
			if nil != vnt {
				vnt.URI = completeURI(urlStr, vnt.URI)
			}
		}

		// 递归
		vnt, e := chooseStream(masterList)
		if nil != e {
			err = e
			return
		}
		return GetPlayList(vnt.URI)
	}
	return
}

func completeURI(siteURI, resourceURI string) string {
	if !strings.HasPrefix(resourceURI, "http") {
		if strings.HasPrefix(resourceURI, "/") {
			// 根目录
			// http://xxx.com/{seg.URI}
			tempUrl := strings.ReplaceAll(siteURI, "://", ":##")
			tempUrl = tempUrl[0:strings.Index(tempUrl, "/")]
			return strings.ReplaceAll(tempUrl, ":##", "://") + resourceURI
		} else {
			// 相对目录
			// http://xxx.com/11/{seg.URI}
			return siteURI[0:strings.LastIndex(siteURI, "/")+1] + resourceURI
		}
	}
	return resourceURI
}

func chooseStream(mediaList *m3u8.MasterPlaylist) (vnt *m3u8.Variant, err error) {
	var sb strings.Builder
	sb.WriteString("Please select program: \n")
	for i, vnt := range mediaList.Variants {
		sb.WriteString(fmt.Sprintf("\t%d. BandWidth: %d, Resolution: %s\n", i+1, vnt.Bandwidth, vnt.Resolution))
	}
	sb.WriteString("Select the number you wanna: ")
	var index int
	fmt.Print(golus.New().FontColor(golus.FontBlue).Apply(sb.String()))
	_, err = fmt.Scan(&index)
	if nil != err {
		return
	}
	if index <= 0 || index > len(mediaList.Variants) {
		err = errors.New("input index out of variants range")
		return
	}
	vnt = mediaList.Variants[index-1]
	fmt.Println(fmt.Sprintf("Your selection is: BandWidth: %d, Resolution: %s", vnt.Bandwidth, vnt.Resolution))
	return
}
