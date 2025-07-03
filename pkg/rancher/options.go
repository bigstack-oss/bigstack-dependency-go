package rancher

var (
	Opts *Options
)

type Option func(*Options)

type Options struct {
	Url  string `yaml:"url"`
	Auth `yaml:"auth"`
}

type Auth struct {
	Token string `yaml:"token"`
}

func Url(url string) Option {
	return func(o *Options) {
		o.Url = url
	}
}

func AuthToken(token string) Option {
	return func(o *Options) {
		o.Auth.Token = token
	}
}
