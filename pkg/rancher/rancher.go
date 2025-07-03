package rancher

import (
	"fmt"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
)

var (
	helper *Helper

	once sync.Once
)

type Helper struct {
	Http    *http.Helper
	Options Options
}

func GetGlobalHelper() *Helper {
	return helper
}

func NewGlobalHelper(opts ...Option) error {
	var err error
	once.Do(func() {
		helper, err = NewHelper(opts...)
	})
	if err != nil {
		return err
	}

	return nil
}

func NewHelper(opts ...Option) (*Helper, error) {
	syncedOpts := syncOptions(opts)
	c, err := http.NewHelper(
		http.TlsInsecureSkipVerify(true),
	)
	if err != nil {
		return nil, err
	}

	return &Helper{
		Http:    c,
		Options: *syncedOpts,
	}, nil
}

func syncOptions(opts []Option) *Options {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	return options
}

func GenAuthHeaders(token string) map[string]string {
	return map[string]string{
		"content-type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
}
