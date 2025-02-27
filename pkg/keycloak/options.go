package keycloak

type Option func(*Options)

type Options struct {
	HostOption `json:"host" yaml:"host"`
	Auth       `json:"auth" yaml:"auth"`
}

type HostOption struct {
	Scheme                string `json:"scheme" yaml:"scheme"`
	Ip                    string `json:"ip" yaml:"ip"`
	Port                  int    `json:"port" yaml:"port"`
	Path                  string `json:"path" yaml:"path"`
	TlsInsecureSkipVerify bool   `json:"tlsInsecureSkipVerify" yaml:"tlsInsecureSkipVerify"`
}

type Auth struct {
	Realm    string `json:"realm" yaml:"realm"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

func Host(scheme string, ip string, port int, path string) Option {
	return func(o *Options) {
		o.Scheme = scheme
		o.Ip = ip
		o.Port = port
		o.Path = path
	}
}

func Insecure(insecure bool) Option {
	return func(o *Options) {
		o.TlsInsecureSkipVerify = insecure
	}
}

func Username(username string) Option {
	return func(o *Options) {
		o.Auth.Username = username
	}
}

func Password(password string) Option {
	return func(o *Options) {
		o.Auth.Password = password
	}
}

func Realm(realm string) Option {
	return func(o *Options) {
		o.Auth.Realm = realm
	}
}
