package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Main string `toml:"main"`
}

const configFileName = "gowt.toml"

func configPath(gitCommonDir string) string {
	return filepath.Join(gitCommonDir, configFileName)
}

func LoadConfig(gitCommonDir string) (Config, error) {
	cfg := Config{}
	path := configPath(gitCommonDir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func SaveConfig(gitCommonDir string, cfg Config) error {
	path := configPath(gitCommonDir)

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
