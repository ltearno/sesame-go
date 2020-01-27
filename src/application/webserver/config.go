package webserver

import (
	"encoding/json"
	"io/ioutil"
)

type ConfigurationFile struct {
	Company              string            `json:"company"`
	IssuerUrl            string            `json:"issuer_url"`
	ListeningAddress     string            `json:"listening_address"`
	TokenDurationSecs    uint64            `json:"token_duration_secs"`
	TokenApiDurationSecs uint64            `json:"token_api_duration_secs"`
	IdTokenDurationSecs  uint64            `json:"id_token_duration_secs"`
	Users                map[string]string `json:"users"`
}

func ReadConfiguration() *ConfigurationFile {
	bytes, err := ioutil.ReadFile("configuration.json")
	checkError(err)

	config := &ConfigurationFile{}
	json.Unmarshal(bytes, &config)

	return config
}
