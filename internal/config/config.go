package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Config représente la structure globale du fichier YAML
type Config struct {
	Programs map[string]ProgramConfig `yaml:"programs"`
}

// ProgramConfig contient tous les paramètres requis pour un job
type ProgramConfig struct {
	Cmd          string            `yaml:"cmd"`
	NumProcs     int               `yaml:"numprocs"`
	Autostart    bool              `yaml:"autostart"`
	Autorestart  string            `yaml:"autorestart"` // "always", "never", "unexpected"
	ExitCodes    []int             `yaml:"exitcodes"`
	StartTime    int               `yaml:"starttime"`
	StartRetries int               `yaml:"startretries"`
	StopSignal   string            `yaml:"stopsignal"`
	StopTime     int               `yaml:"stoptime"`
	Stdout       string            `yaml:"stdout"`
	Stderr       string            `yaml:"stderr"`
	Env          map[string]string `yaml:"env"`
	WorkingDir   string            `yaml:"workingdir"`
	Umask        string            `yaml:"umask"`
}

// LoadConfig lit le fichier YAML et le transforme en structure Go
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
