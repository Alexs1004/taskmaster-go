package daemon

import (
	"os/exec"
	// "syscall"
	"time"
	
	"github.com/Alexs1004/taskmaster-go/internal/config"
)

// Process représente un processus managé par Taskmaster
type Process struct {
	Name      string
	Config    config.ProgramConfig
	Cmd       *exec.Cmd
	State     string    // "STOPPED", "STARTING", "RUNNING", "BACKOFF", "EXITED"
	Pid       int
	StartedAt time.Time
}

// NewProcess initialise une nouvelle instance de processus sans la lancer
func NewProcess(name string, cfg config.ProgramConfig) *Process {
	return &Process{
		Name:   name,
		Config: cfg,
		State:  "STOPPED",
	}
}

// Start lance le processus en arrière-plan (équivalent de fork/exec)
func (p *Process) Start() error {
	// On prépare la commande
	p.Cmd = exec.Command("sh", "-c", p.Config.Cmd)
	
	// Si un dossier de travail est spécifié, on l'applique
	if p.Config.WorkingDir != "" {
		p.Cmd.Dir = p.Config.WorkingDir
	}

	// C'est ici que Go gère le fork/exec en arrière-plan sans bloquer
	err := p.Cmd.Start()
	if err != nil {
		p.State = "BACKOFF"
		return err
	}

	// On met à jour l'état de notre machine à états interne
	p.Pid = p.Cmd.Process.Pid
	p.State = "STARTING"
	p.StartedAt = time.Now()

	return nil
}
