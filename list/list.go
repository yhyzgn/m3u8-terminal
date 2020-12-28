// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-28 17:51
// version: 1.0.0
// desc   : 

package list

import (
	"bytes"
	"m3u8/http"
	"m3u8/m3u8"
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
			}
		}
	} else {
		// master
		masterList = playList.(*m3u8.MasterPlaylist)
		for _, vnt := range masterList.Variants {
			if nil != vnt {
				vnt.URI = completeURI(urlStr, vnt.URI)
			}
		}
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
