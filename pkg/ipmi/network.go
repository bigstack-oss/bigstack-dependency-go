package ipmi

import (
	"fmt"
	"strings"
)

func (h *Helper) parseIpmiIp(out []byte) (string, error) {
	lines := strings.SplitSeq(string(out), "\n")

	for line := range lines {
		if !strings.HasPrefix(line, "IP Address") {
			continue
		}

		if strings.Contains(line, "IP Address Source") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			ip := strings.TrimSpace(parts[1])
			return ip, nil
		}
	}

	return "", fmt.Errorf("ipmi ip not found in output")
}
