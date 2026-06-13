package config

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

// Config représente l'arborescence racine du fichier de configuration YAML.
type Config struct {
	Programs map[string]ProgramConfig `yaml:"programs"`
}

// ProgramConfig contient l'intégralité des attributs d'exécution d'un programme.
type ProgramConfig struct {
	Cmd          string            `yaml:"cmd"`
	NumProcs     int               `yaml:"numprocs"`
	Autostart    bool              `yaml:"autostart"`
	Autorestart  string            `yaml:"autorestart"` // Valeurs acceptées : "always", "never", "unexpected"
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

// LoadConfig extrait le contenu d'un fichier YAML, le déserialize et valide la cohérence des données.
func LoadConfig(path string) (*Config, error) {
	// Lecture brute du fichier sur le disque
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Déserialization du conteneur YAML vers la structure Go
	var cfg Config
	 // Unmarshal convertit le YAML en structure Go grâce aux tags `yaml` définis dans les structs
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	// --- Phase d'assainissement et de validation des données ---
	for name, prog := range cfg.Programs {
		// Validation de 'numprocs' : si absent ou incorrect, initialisation obligatoire à 1 instance
		if prog.NumProcs <= 0 {
			prog.NumProcs = 1
		}
		
		// Validation de la commande : un programme sans binaire ou script associé doit invalider la configuration
		if prog.Cmd == "" {
			return nil, fmt.Errorf("erreur de configuration : le programme '%s' n'a pas de commande (cmd) définie", name)
		}

		// Réinjection de la structure modifiée (les structures sont passées par valeur dans les maps)
		cfg.Programs[name] = prog
	}

	return &cfg, nil
}