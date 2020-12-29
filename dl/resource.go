// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-28 16:45
// version: 1.0.0
// desc   : 

package dl

type Resource struct {
	index     int
	URL       string
	Filename  string
	Overwrite bool
}

func NewResource(url, filename string, overwrite bool) *Resource {
	return &Resource{
		URL:       url,
		Filename:  filename,
		Overwrite: overwrite,
	}
}
