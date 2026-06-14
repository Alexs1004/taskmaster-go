package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Alexs1004/taskmaster-go/internal/daemon"
)

func RunShell(pm *daemon.ProcessManager) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("taskmaster> ")

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		clearLine := strings.TrimSpace(line)
		parts := strings.Fields(clearLine)
		
		if len(parts) == 0 {
			continue
		}
		
		cmd := parts[0]

		switch cmd {
		case "status":
			if len(parts) > 1 {
				// L'utilisateur a tapé "status my_worker_0"
				pm.SpecificProcessStatus(parts[1])
			} else {
				// L'utilisateur a juste tapé "status"
				pm.Status()
			}

		case "start":
			if len(parts) > 1 {
				pm.StartProcess(parts[1])
			} else {
				fmt.Println("Erreur: veuillez spécifier un processus (ex: start my_worker_0)")
			}

		case "stop":
			if len(parts) > 1 {
				err := pm.StopProcess(parts[1])
				if err != nil {
					fmt.Printf("Erreur : %v\n", err)
				} else {
					fmt.Printf("Signal d'arrêt envoyé à %s\n", parts[1])
				}
			} else {
				fmt.Println("Erreur: veuillez spécifier un processus (ex: stop my_worker_0)")
			}

		case "exit", "quit":
			fmt.Println("Fermeture du Taskmaster. Arrêt des processus en cours...")
			pm.StopAllProcesses() 
			return

		default:
			fmt.Printf("Commande inconnue : %s. Commandes disponibles : status, start, stop, exit\n", cmd)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Erreur de lecture :", err)
	}
}