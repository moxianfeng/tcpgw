package method

import "github.com/moxianfeng/gofactory"

type None struct {
	MethodBase
}

func (this *None) Process(in []byte) ([]byte, error) {
	return in, nil
}

func init() {
	gofactory.Default.Register("none", &None{})
	gofactory.Default.Register("", &None{})
}
