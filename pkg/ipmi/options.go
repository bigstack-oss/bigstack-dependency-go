package ipmi

const (
	defaultPort = 623
)

type Option func(*Options)

type Options struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func Host(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
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
