package daemon

import (
	"fmt"
	"sync"
	"os"
	"text/tabwriter"
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

	for _, proc := range pm.Processes {
		err := proc.Start()
		if err != nil {
			// Enregistrement de l'erreur sans bloquer le démarrage des autres programmes
			fmt.Printf("Erreur au démarrage de %s: %v\n", proc.Name, err)
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