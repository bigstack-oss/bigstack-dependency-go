package ipmi

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

var (
	cmd = exec.Command
)

type Helper struct {
	cancel *context.CancelFunc
	Options
}

func initOptions(opts []Option) *Options {
	options := genDefaultOptions()
	for _, o := range opts {
		o(options)
	}

	return options
}

func genDefaultOptions() *Options {
	return &Options{Port: defaultPort}
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)
	return &Helper{
		Options: *initedOpts,
	}, nil
}

func (h *Helper) GetFRU() (*FRU, error) {
	out, err := cmd("ipmitool", "-I", "lanplus", "-H", h.Host, "-p", strconv.Itoa(h.Port), "-U", h.Username, "-P", h.Password, "fru", "print", "0").Output()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get IPMI fru for 0 %s(%v)",
			err,
			string(out),
		)
	}

	return h.parseFRU(out)
}

func (h *Helper) GetDefaultIpmiIp() (string, error) {
	out, err := cmd("ipmitool", "-I", "lanplus", "-H", h.Host, "-p", strconv.Itoa(h.Port), "-U", h.Username, "-P", h.Password, "lan", "print", "1").Output()
	if err != nil {
		return "", fmt.Errorf(
			"failed to get IPMI ip %s(%v)",
			string(out),
			err,
		)
	}

	return h.parseIpmiIp(out)
}

func (h *Helper) Operate(operation string) error {
	out, err := cmd("ipmitool", "-I", "lanplus", "-H", h.Host, "-p", strconv.Itoa(h.Port), "-U", h.Username, "-P", h.Password, "chassis", "power", operation).Output()
	if err != nil {
		return fmt.Errorf(
			"failed to do IPMI operation %s(%v)",
			string(out),
			err,
		)
	}

	return nil
}
