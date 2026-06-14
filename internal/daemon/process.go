package daemon

import (
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/Alexs1004/taskmaster-go/internal/config"
)

// Process représente un processus managé par Taskmaster
type Process struct {
	Name          string
	Config        config.ProgramConfig
	Cmd           *exec.Cmd
	State         string // "STOPPED", "STARTING", "RUNNING", "BACKOFF", "EXITED", "FATAL"
	Pid           int
	StartedAt     time.Time
	RetriesCount  int
	SuccessTimer  *time.Timer
	IsIntentional bool
	mu            sync.Mutex // Mutex pour protéger les accès concurrents
}

// NewProcess initialise une nouvelle instance de processus sans la lancer
func NewProcess(name string, cfg config.ProgramConfig) *Process {
	return &Process{
		Name:   name,
		Config: cfg,
		State:  "STOPPED",
	}
}

// Start lance le processus en arrière-plan et active sa surveillance
func (p *Process) Start() error {
	p.mu.Lock()

	if p.State == "RUNNING" || p.State == "STARTING" {
		p.mu.Unlock()
		return fmt.Errorf("le processus est déjà en cours d'exécution (état: %s)", p.State)
	}

	p.IsIntentional = false // Protégé car partagé avec le CLI (futur "stop")
	p.Cmd = exec.Command("sh", "-c", p.Config.Cmd)
	if p.Config.WorkingDir != "" {
		p.Cmd.Dir = p.Config.WorkingDir
	}
	p.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Création d'un groupe de processus pour permettre de tuer le processus et ses enfants
	p.mu.Unlock()

	err := p.Cmd.Start()
	if err != nil {
		p.mu.Lock()
		p.State = "BACKOFF"
		p.mu.Unlock()
		return err
	}

	p.mu.Lock()
	p.Pid = p.Cmd.Process.Pid
	p.State = "STARTING"
	p.StartedAt = time.Now()
	p.mu.Unlock()

	// 1. Lancement du Timer de succès
	p.SuccessTimer = time.AfterFunc(time.Duration(p.Config.StartTime)*time.Second, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.State == "STARTING" {
			p.State = "RUNNING"
			p.RetriesCount = 0
		}
	})

	// 2. Déclenchement du gardien
	go p.monitor()

	return nil
}

// Stop tente d'arrêter le processus de manière intentionnelle, en tuant le processus enfant
func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.State == "STOPPED" || p.State == "EXITED" || p.State == "FATAL" {
		return nil
	}

	p.IsIntentional = true
	if p.Cmd != nil && p.Cmd.Process != nil {
		err := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
		if err != nil {
			if err.Error() == "no such process" {
				return nil
			}
			return err
		}
	}
	return nil
}

// monitor attend la fin du processus et applique la logique de redémarrage
func (p *Process) monitor() {
	err := p.Cmd.Wait()

	if p.SuccessTimer != nil {
		p.SuccessTimer.Stop()
	}

	// Protection de la lecture d'IsIntentional et modification de l'état
	p.mu.Lock()
	if p.IsIntentional {
		p.State = "STOPPED"
		p.Pid = 0
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	// Extraction du code de sortie
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			// Protection de la modification en cas d'erreur système
			p.mu.Lock()
			p.State = "BACKOFF"
			p.Pid = 0
			p.mu.Unlock()
			return
		}
	}

	isExpected := false
	for _, expected := range p.Config.ExitCodes {
		if exitCode == expected {
			isExpected = true
			break
		}
	}

	// Gestion finale de la machine à états
	p.mu.Lock()
	p.Pid = 0
	if isExpected && p.Config.Autorestart != "always" {
		p.State = "EXITED"
		p.mu.Unlock()
	} else if p.Config.Autorestart == "never" {
		p.State = "EXITED"
		p.mu.Unlock()
	} else {
		p.mu.Unlock() // On déverrouille AVANT d'appeler handleRestart qui va reverrouiller
		p.handleRestart()
	}
}

// handleRestart gère les tentatives de relance en fonction de la configuration
func (p *Process) handleRestart() {
	p.mu.Lock()
	if p.RetriesCount >= p.Config.StartRetries {
		p.State = "FATAL"
		p.mu.Unlock()
		return
	}

	p.RetriesCount++
	p.State = "BACKOFF"
	p.mu.Unlock()

	time.Sleep(1 * time.Second)
	_ = p.Start()
}

// GetStatusInfo renvoie une copie des données actuelles en toute sécurité
func (p *Process) GetStatusInfo() (int, string, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Pid, p.State, p.RetriesCount
}