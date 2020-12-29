// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-29 14:36
// version: 1.0.0
// desc   : 

package settings

type Config struct {
	SaveDir         string `json:"saveDir"`
	TsTempDirPrefix string `json:"tsTempDirPrefix"`
	Ffmpeg          struct {
		Windows string `json:"windows"`
		Mac     string `json:"mac"`
		Linux   string `json:"linux"`
	} `json:"ffmpeg"`
}
