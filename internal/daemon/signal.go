package daemon

import (
	"strings"
	"syscall"
)

// ParseSignal convertit une chaîne de caractères issue du YAML en un signal UNIX.
func ParseSignal(sigStr string) syscall.Signal {
	cleanSig := strings.ToUpper(strings.TrimSpace(sigStr))

	switch cleanSig {
	case "TERM", "SIGTERM":
		return syscall.SIGTERM
	case "INT", "SIGINT":
		return syscall.SIGINT
	case "QUIT", "SIGQUIT":
		return syscall.SIGQUIT
	case "KILL", "SIGKILL":
		return syscall.SIGKILL
	case "HUP", "SIGHUP":
		return syscall.SIGHUP
	case "USR1", "SIGUSR1":
		return syscall.SIGUSR1
	case "USR2", "SIGUSR2":
		return syscall.SIGUSR2
	default:
		return syscall.SIGTERM
	}
}