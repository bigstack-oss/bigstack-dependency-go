package rancher

import "encoding/json"

type Secret struct {
	Type     string `json:"type"`
	Metadata `json:"metadata"`
	Data     `json:"data"`
}

type Data struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SecretResponse struct {
	Metadata `json:"metadata"`
}

func (s *Secret) Bytes() ([]byte, error) {
	return json.Marshal(s)
}
