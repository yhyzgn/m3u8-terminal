// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-28 16:45
// version: 1.0.0
// desc   : 

package dl

type Resource struct {
	index    int
	URL      string
	Filename string
}

func NewResource(url, filename string) *Resource {
	return &Resource{
		URL:      url,
		Filename: filename,
	}
}
