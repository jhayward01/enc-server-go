package utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

func VerifyConfigs(configs map[string]string, expected []string) (ok bool, missing string) {

	// Iterate across expected values in config map.
	for _, s := range expected {
		if _, ok = configs[s]; !ok {
			return false, s
		}
	}
	return true, ""
}

func LoadConfigs(configPath string) (configs map[string]map[string]string, err error) {

	// Load config path overrides.
	if val, ok := os.LookupEnv("ENC_SERVER_GO_CONFIG_PATH"); ok {
		configPath = val
	}

	// Load YAML file.
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var configEntries map[string]map[string]string
	if err = yaml.Unmarshal(yamlFile, &configEntries); err != nil {
		return nil, err
	}

	return configEntries, nil
}
