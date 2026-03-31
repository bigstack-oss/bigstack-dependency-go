package terraform

var (
	Opts *Options
)

type Option func(*Options)

type Options struct {
	Version   string
	WoringDir string
	ExecPath  string
}

func Version(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

func WorkingDir(dir string) Option {
	return func(o *Options) {
		o.WoringDir = dir
	}
}

func ExecPath(path string) Option {
	return func(o *Options) {
		o.ExecPath = path
	}
}
