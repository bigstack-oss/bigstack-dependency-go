package mongo

var (
	Opts *Options
)

type Option func(*Options)

type Options struct {
	Uri        string `json:"uri" yaml:"uri"`
	Host       string `json:"host" yaml:"host"`
	Port       int    `json:"port" yaml:"port"`
	Auth       `json:"auth" yaml:"auth"`
	ReplicaSet string `json:"replicaSet" yaml:"replicaSet"`
	Connect    string `json:"connect" yaml:"connect"`

	Database    string            `json:"database" yaml:"database"`
	Collection  string            `json:"collection" yaml:"collection"`
	Databases   map[string]string `json:"databases" yaml:"databases"`
	Collections map[string]string `json:"collections" yaml:"collections"`
}

type Auth struct {
	Enable   bool   `json:"enable" yaml:"enable"`
	Source   string `json:"source" yaml:"source"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

func Uri(uri string) Option {
	return func(o *Options) {
		o.Uri = uri
	}
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

func ReplicaSet(replicaSet string) Option {
	return func(o *Options) {
		o.ReplicaSet = replicaSet
	}
}

func Connect(connect string) Option {
	return func(o *Options) {
		o.Connect = connect
	}
}

func Database(database string) Option {
	return func(o *Options) {
		o.Database = database
	}
}

func Collection(collection string) Option {
	return func(o *Options) {
		o.Collection = collection
	}
}

func Databases(databases map[string]string) Option {
	return func(o *Options) {
		o.Databases = databases
	}
}

func Collections(collections map[string]string) Option {
	return func(o *Options) {
		o.Collections = collections
	}
}

func AuthEnable(enable bool) Option {
	return func(o *Options) {
		o.Auth.Enable = enable
	}
}

func AuthSource(source string) Option {
	return func(o *Options) {
		o.Auth.Source = source
	}
}

func AuthUsername(username string) Option {
	return func(o *Options) {
		o.Auth.Username = username
	}
}

func AuthPassword(password string) Option {
	return func(o *Options) {
		o.Auth.Password = password
	}
}
