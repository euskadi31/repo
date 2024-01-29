package config

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

var (
	repoDirName    = ".repo"
	repoConfigName = "repo.yml"
	repoDir        string
)

type Config struct {
	Repos []string `yaml:"repos" json:"repos"`
}

func GetRepoDir() (string, error) {
	if repoDir != "" {
		return repoDir, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get user home directory: %w", err)
	}

	repoDir = path.Join(homeDir, repoDirName)

	return repoDir, nil
}

func Init() error {
	repoDir, err := GetRepoDir()
	if err != nil {
		return fmt.Errorf("cannot get repo directory: %w", err)
	}

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		if err := os.MkdirAll(repoDir, os.ModePerm); err != nil {
			return fmt.Errorf("cannot create repo directory %s: %w", repoDir, err)
		}
	}

	if _, err := os.Stat(path.Join(repoDir, repoConfigName)); os.IsExist(err) {
		return nil
	}

	cfg := &Config{
		Repos: []string{},
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(path.Join(repoDir, repoConfigName), b, os.ModePerm); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}

	return nil
}

func Read() (*Config, error) {
	repoDir, err := GetRepoDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get repo directory: %w", err)
	}

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("please run repo init")
	}

	b, err := os.ReadFile(path.Join(repoDir, repoConfigName))
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config file: %w", err)
	}

	return cfg, nil
}
