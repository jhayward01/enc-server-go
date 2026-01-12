package utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Configs struct {
	FeClientConfigs FeClientConfigs `yaml:"feClientConfigs"`
	FeServerConfigs FeServerConfigs `yaml:"feServerConfigs"`
	BeServerConfigs BeServerConfigs `yaml:"beServerConfigs"`
	BeClientConfigs BeClientConfigs `yaml:"beClientConfigs"`
	TestParams      TestParams      `yaml:"testParams"`
}

type FeClientConfigs struct {
	ServerAddr string `yaml:"serverAddr"`
}

type FeServerConfigs struct {
	KeySize    string `yaml:"keySize"`
	IdKeyStr   string `yaml:"idKeyStr"`
	IdNonceStr string `yaml:"idNonceStr"`
	Port       string `yaml:"port"`
}

type BeServerConfigs struct {
	Port     string `yaml:"port"`
	MongoURI string `yaml:"mongoURI"`
}

type BeClientConfigs struct {
	ServerAddr string `yaml:"serverAddr"`
}

type TestParams struct {
	Id     string `yaml:"id"`
	Record string `yaml:"record"`
}

func VerifyTopConfigs(configs map[string]map[string]string, expected []string) (ok bool, missing string) {

	// Iterate across expected values in config map.
	for _, s := range expected {
		if _, ok = configs[s]; !ok {
			return false, s
		}
	}
	return true, ""
}

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

func LoadConfigs2(configPath string) (configs *Configs, err error) {

	// Load config path overrides.
	if val, ok := os.LookupEnv("ENC_SERVER_GO_CONFIG_PATH"); ok {
		configPath = val
	}

	// Load YAML file.
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Unmarshal YAML into the struct
	if err := yaml.Unmarshal(yamlFile, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}
