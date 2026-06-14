package daemon

import (
	"fmt"
	"sync"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Alexs1004/taskmaster-go/internal/config"
)

// ProcessManager orchestre et centralise l'ensemble des processus managés.
type ProcessManager struct {
	Processes map[string]*Process // Dictionnaire associant un nom unique à son instance de processus
	mu        sync.RWMutex        // Mutex de lecture/écriture pour sécuriser l'accès concurrent à la map
}

// NewProcessManager alloue et initialise un gestionnaire de processus prêt à l'emploi.
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		// Initialisation obligatoire de la map pour éviter un panic lors des écritures
		Processes: make(map[string]*Process),
	}
}

// StartAllProcesses parcourt et démarre chaque processus enregistré de manière asynchrone.
func (pm *ProcessManager) StartAllProcesses() {
	// RLock permet à plusieurs Goroutines de lire simultanément mais bloque toute écriture concurrente
	pm.mu.RLock() 
	defer pm.mu.RUnlock() // Garantie absolue de libération du verrou à la sortie de la méthode

	for _, p := range pm.Processes {
		if p.Config.Autostart {
			err := p.Start()
			if err != nil {
				fmt.Printf("Erreur au démarrage de %s: %v\n", p.Name, err)
			}
		} else {
			// S'il ne doit pas démarrer, on s'assure que son état initial est correct
			p.mu.Lock()
			p.State = "STOPPED"
			p.mu.Unlock()
		}
	}
}

// AddProcess ajoute de manière thread-safe un nouveau processus dans le dictionnaire du gestionnaire.
func (pm *ProcessManager) AddProcess(proc *Process) {
	// Lock exclusif : interdit toute lecture ou écriture concurrente pendant la modification de la map
	pm.mu.Lock() 
	defer pm.mu.Unlock() // Garantie absolue de libération du verrou à la sortie de la méthode
	
	// Enregistrement de l'instance avec son nom unique comme clé d'accès
	pm.Processes[proc.Name] = proc
}

func (pm *ProcessManager) Status() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Utilisation du Tabwriter pour un rendu professionnel
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tPID\tSTATE\tRETRIES")

	for name, proc := range pm.Processes {
		pid, state, retries := proc.GetStatusInfo()
		
		fmt.Fprintf(w, "%s\t%d\t%s\t%d\n", name, pid, state, retries)
	}
	w.Flush()
}

func (pm *ProcessManager) SpecificProcessStatus(name string) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	proc, exists := pm.Processes[name]
	if !exists {
		fmt.Printf("Processus %s non trouvé.\n", name)
		return
	}

	pid, state, retries := proc.GetStatusInfo()
	fmt.Printf("Statut du processus %s: PID: %d, State: %s, Retries: %d\n", 
		name, pid, state, retries)
}

func (pm *ProcessManager) StartProcess(name string) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	proc, exists := pm.Processes[name]
	if !exists {
		return fmt.Errorf("processus %s non trouvé", name)
	}

	return proc.Start()
}

func (pm *ProcessManager) StopProcess(name string) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	proc, exists := pm.Processes[name]
	if !exists {
		return fmt.Errorf("processus %s non trouvé", name)
	}
	return proc.Stop()
}

func (pm *ProcessManager) StopAllProcesses() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, proc := range pm.Processes {
		err := proc.Stop()
		if err != nil {
			fmt.Printf("Erreur à l'arrêt de %s: %v\n", proc.Name, err)
		}
	}
}

// ReloadConfig lit le fichier YAML et applique les modifications à chaud.
func (pm *ProcessManager) ReloadConfig(configPath string) error {
	newCfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("erreur de lecture du YAML: %v", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Liste temporaire pour repérer les instances valides dans le nouveau YAML
	expectedProcs := make(map[string]bool)
	for progName, progCfg := range newCfg.Programs {
		for i := 0; i < progCfg.NumProcs; i++ {
			expectedProcs[fmt.Sprintf("%s_%d", progName, i)] = true
		}
	}

	// Phase A : Nettoyage (Arrêt des processus supprimés, réduits, ou modifiés)
	for procName, proc := range pm.Processes {
		// Extraction du nom de base (ex: "my_worker" depuis "my_worker_0")
		baseName := procName[:strings.LastIndex(procName, "_")]
		newProgCfg, exists := newCfg.Programs[baseName]

		needsStop := false
		if !exists {
			needsStop = true // Le bloc entier a disparu du YAML
		} else if !expectedProcs[procName] {
			needsStop = true // Le numprocs a été réduit
		} else if proc.Config.HasChanged(newProgCfg) {
			needsStop = true // Une option a été modifiée
		}

		if needsStop {
			fmt.Printf("[Reload] Arrêt du processus obsolète ou modifié : %s\n", procName)
			_ = proc.Stop()
			delete(pm.Processes, procName) // Suppression définitive de la mémoire
		}
	}

	// Phase B : Démarrage (Ajout des nouveaux processus ou de ceux qui ont été recréés)
	for progName, newProgCfg := range newCfg.Programs {
		for i := 0; i < newProgCfg.NumProcs; i++ {
			procName := fmt.Sprintf("%s_%d", progName, i)

			// Si le processus n'est plus dans notre map, c'est qu'il est nouveau ou fraîchement nettoyé
			if _, exists := pm.Processes[procName]; !exists {
				fmt.Printf("[Reload] Configuration appliquée pour : %s\n", procName)
				newProc := NewProcess(procName, newProgCfg)
				pm.Processes[procName] = newProc

				if newProgCfg.Autostart {
					err := newProc.Start()
					if err != nil {
						fmt.Printf("[Reload] Erreur au démarrage de %s: %v\n", procName, err)
					}
				}
			}
		}
	}

	return nil
}