package daemon

import (
	"fmt"
	"sync"
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