package aws

import (
	"context"
	"sync"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	RegionAuto = "auto"

	S3DefaultAccessKey = ""
	S3DefaultSecretKey = ""
)

var (
	helper *Helper
	Opts   *Options
	once   sync.Once
)

type S3Client interface {
	CreateBucket(context.Context, *s3.CreateBucketInput, ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type S3PresignedClient interface {
	PresignGetObject(context.Context, *s3.GetObjectInput, ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type Helper struct {
	AwsS3Client *s3.Client

	S3Client
	S3PresignedClient

	Options
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
	initedOpts := initOptions(opts)
	h := &Helper{Options: *initedOpts}

	err := h.SetS3Client()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func initOptions(opts []Option) *Options {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	return options
}
