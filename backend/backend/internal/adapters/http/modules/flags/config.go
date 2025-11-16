package flags

import (
	"os"

	"gopkg.in/yaml.v3"
)

// HeaderConfig maps HTTP header names to flag names.
type HeaderConfig map[string]string

const defaultConfigPath = "config/headers.yml"

// LoadHeaderConfigFrom loads header flag configuration from the given path.
// Missing files result in an empty configuration.
func LoadHeaderConfigFrom(path string) (HeaderConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return HeaderConfig{}, nil
		}
		return nil, err
	}
	cfg := HeaderConfig{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadHeaderConfig loads header flag configuration from environment variable
// FLAGS_HEADER_PATH or defaults to config/headers.yml.
func LoadHeaderConfig() (HeaderConfig, error) {
	path := os.Getenv("FLAGS_HEADER_PATH")
	if path == "" {
		path = defaultConfigPath
	}
	return LoadHeaderConfigFrom(path)
}
