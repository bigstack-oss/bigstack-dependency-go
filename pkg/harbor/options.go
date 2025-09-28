package harbor

var (
	Opts *Options
)

type Option func(*Options)

type Options struct {
	Url                string
	Username           string
	Password           string
	InsecureSkipVerify bool
}

func Url(url string) Option {
	return func(o *Options) {
		o.Url = url
	}
}

func Username(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

func Password(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

func InsecureSkipVerify(enabled bool) Option {
	return func(o *Options) {
		o.InsecureSkipVerify = enabled
	}
}
