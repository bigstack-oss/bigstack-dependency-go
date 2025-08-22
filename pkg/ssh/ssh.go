package ssh

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/sftp"
	log "go-micro.dev/v5/logger"
	"golang.org/x/crypto/ssh"
)

var (
	helper *Helper
	once   sync.Once
)

type DialClient interface {
	NewSession() (*ssh.Session, error)
	Close() error
}

type SessionClient interface {
	Run(string) error
	Close() error
}

type SftpClient interface {
	Create(string) (*sftp.File, error)
}

type Helper struct {
	SshClientPtr *ssh.Client
	DialClient
	SessionClient
	SftpClient
	Options
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)
	h := &Helper{Options: *initedOpts}

	err := h.SetClients()
	if err != nil {
		return nil, err
	}

	return h, nil
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

func GetGlobalHelper() *Helper {
	return helper
}

func initOptions(opts []Option) *Options {
	options := &Options{Timeout: 30 * time.Second}
	for _, o := range opts {
		o(options)
	}

	return options
}

func (h *Helper) SetClients() error {
	config := &ssh.ClientConfig{
		User:            h.User,
		HostKeyCallback: h.HostKeyCallback,
		Timeout:         h.Timeout,
	}

	var err error
	h.SshClientPtr, err = ssh.Dial("tcp", h.Host, config)
	if err != nil {
		log.Errorf("ssh: failed to connect to node %s(%v)", h.Host, err)
		return err
	}

	h.DialClient = h.SshClientPtr
	h.SessionClient, err = h.DialClient.NewSession()
	if err != nil {
		log.Errorf("ssh: failed to create session for node %s(%v)", h.Host, err)
		return err
	}

	h.SftpClient, err = sftp.NewClient(h.SshClientPtr)
	if err != nil {
		log.Errorf("ssh: failed to create SFTP client for node %s(%v)", h.Host, err)
		return err
	}

	return nil
}

func (h *Helper) Run(cmd string) error {
	if h.SessionClient == nil {
		return fmt.Errorf("session client is not initialized")
	}

	err := h.SessionClient.Run(cmd)
	if err != nil {
		log.Errorf("ssh: failed to run command %s on node %s(%v)", cmd, h.Host, err)
		return err
	}

	return nil
}

func (h *Helper) Copy(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.Errorf("ssh: failed to open source file %s(%v)", srcPath, err)
		return err
	}

	defer srcFile.Close()
	dstFile, err := h.SftpClient.Create(dstPath)
	if err != nil {
		log.Errorf("ssh: failed to create destination file %s(%v)", dstPath, err)
		return err
	}

	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Errorf("ssh: failed to copy file from %s to %s:%s(%v)", srcPath, h.Host, dstPath, err)
		return err
	}

	return nil
}

func (h *Helper) Close() {
	if h.SessionClient != nil {
		h.SessionClient.Close()
	}

	if h.DialClient != nil {
		h.DialClient.Close()
	}

	if h.SshClientPtr != nil {
		h.SshClientPtr.Close()
	}
}
