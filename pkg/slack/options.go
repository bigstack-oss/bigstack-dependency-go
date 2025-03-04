package slack

type Option func(*Options)

type Options struct {
	Token string
}

func Token(token string) Option {
	return func(o *Options) {
		o.Token = token
	}
}
