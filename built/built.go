// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-30 9:45
// version: 1.0.0
// desc   : 

package built

import "fmt"

const (
	name = "m3u8 downloader"
)

var (
	Version   = "dev"
	BuildDate = "now"
)

var (
	FullName string
)

func init() {
	FullName = fmt.Sprintf("%s version \"%s\" %s\n", name, Version, BuildDate)
}
