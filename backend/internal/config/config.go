package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Mmctl    MmctlConfig    `yaml:"mmctl"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type MmctlConfig struct {
	BaseDir     string `yaml:"base_dir"`
	ScriptPath  string `yaml:"script_path"`
	WorkDir     string `yaml:"work_dir"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Output == "" {
		cfg.Log.Output = "stdout"
	}
	if cfg.Mmctl.BaseDir == "" {
		cfg.Mmctl.BaseDir = "/data"
	}
	if cfg.Mmctl.ScriptPath == "" {
		cfg.Mmctl.ScriptPath = "/root/.openclaw/workspace/repos/helm_maigc2/mmctl.sh"
	}
	if cfg.Mmctl.WorkDir == "" {
		cfg.Mmctl.WorkDir = filepath.Join(cfg.Mmctl.BaseDir, "clusters")
	}

	return &cfg, nil
}
