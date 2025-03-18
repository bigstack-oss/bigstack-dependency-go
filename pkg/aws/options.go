package aws

type Option func(*Options)

type Options struct {
	Region string `json:"region" yaml:"region"`
	S3Url  string `json:"s3Url" yaml:"s3Url"`

	AccessKey string `json:"accessKey" yaml:"accessKey"`
	SecretKey string `json:"secretKey" yaml:"secretKey"`

	EnableCustomURL    bool `json:"enableCustomURL" yaml:"enableCustomURL"`
	EnableStaticCreds  bool `json:"enableStaticCreds" yaml:"enableStaticCreds"`
	InsecureSkipVerify bool `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
}

func Region(region string) Option {
	return func(o *Options) {
		o.Region = region
	}
}

func S3Url(s3Url string) Option {
	return func(o *Options) {
		o.S3Url = s3Url
	}
}

func AccessKey(accessKey string) Option {
	return func(o *Options) {
		o.AccessKey = accessKey
	}
}

func SecretKey(secretKey string) Option {
	return func(o *Options) {
		o.SecretKey = secretKey
	}
}

func EnableCustomURL(enable bool) Option {
	return func(o *Options) {
		o.EnableCustomURL = enable
	}
}

func EnableStaticCreds(enable bool) Option {
	return func(o *Options) {
		o.EnableStaticCreds = enable
	}
}

func InsecureSkipVerify(skip bool) Option {
	return func(o *Options) {
		o.InsecureSkipVerify = skip
	}
}
