package daemon

import (
	"os/exec"
	"syscall"
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
	RetriesCount int
	SuccessTimer *time.Timer   // Timer pour valider le passage à "RUNNING"
	IsIntentional bool         // Flag pour savoir si l'arrêt vient d'un ordre utilisateur
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
	p.Cmd = exec.Command("sh", "-c", p.Config.Cmd)
	p.IsIntentional = false // Reset du flag au démarrage

	if p.Config.WorkingDir != "" {
		p.Cmd.Dir = p.Config.WorkingDir
	}

	err := p.Cmd.Start()
	if err != nil {
		p.State = "BACKOFF"
		return err
	}

	p.Pid = p.Cmd.Process.Pid
	p.State = "STARTING"
	p.StartedAt = time.Now()

	// 1. Lancement du Timer de succès (StartTime)
	// Au bout du temps requis, si l'état est toujours STARTING, on valide le RUNNING
	p.SuccessTimer = time.AfterFunc(time.Duration(p.Config.StartTime)*time.Second, func() {
		if p.State == "STARTING" {
			p.State = "RUNNING"
			p.RetriesCount = 0 // Succès total : on réinitialise le droit à l'erreur
		}
	})

	// 2. Déclenchement du gardien en arrière-plan via une Goroutine
	go p.monitor()

	return nil
}

// monitor attend la fin du processus et applique la logique de redémarrage
func (p *Process) monitor() {
	// On attend la fin effective du processus (bloquant dans cette Goroutine)
	err := p.Cmd.Wait()

	// Dès qu'on arrive ici, le processus est mort.
	// Si un timer de succès tournait encore, on l'annule immédiatement
	if p.SuccessTimer != nil {
		p.SuccessTimer.Stop()
	}

	// Si l'arrêt est voulu par l'utilisateur (via une future commande stop), on stoppe la boucle
	if p.IsIntentional {
		p.State = "STOPPED"
		p.Pid = 0
		return
	}

	// On extrait le code de sortie (Exit Code)
	exitCode := 0
	if err != nil {
		// Si err n'est pas nil, on essaie de récupérer le code de retour UNIX
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			// Erreur système inattendue
			p.State = "BACKOFF"
			p.Pid = 0
			return
		}
	}

	// On vérifie si le code de sortie est considéré comme un succès (attendu)
	isExpected := false
	for _, expected := range p.Config.ExitCodes {
		if exitCode == expected {
			isExpected = true
			break
		}
	}

	// Gestion de la machine à états et de l'Auto-restart
	p.Pid = 0
	if isExpected && p.Config.Autorestart != "always" {
		// Le programme s'est arrêté proprement, et on ne demande pas de le relancer "toujours"
		p.State = "EXITED"
	} else if p.Config.Autorestart == "never" {
		// Crash ou arrêt propre, mais la politique interdit le redémarrage
		p.State = "EXITED"
	} else {
		// C'est un crash inattendu OU la politique est réglée sur "always"
		p.handleRestart()
	}
}

// handleRestart gère les tentatives de relance en fonction de la configuration
func (p *Process) handleRestart() {
	if p.RetriesCount >= p.Config.StartRetries {
		// On a épuisé toutes nos chances
		p.State = "FATAL"
		return
	}

	// On incrémente le compteur et on bascule en BACKOFF avant de relancer
	p.RetriesCount++
	p.State = "BACKOFF"

	// Petite pause de sécurité pour éviter de surcharger le CPU en cas de boucle folle
	time.Sleep(1 * time.Second)

	// On relance le processus !
	_ = p.Start()
}

