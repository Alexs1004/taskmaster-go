# Taskmaster

A lightweight, robust, and concurrent UNIX process manager written in Go, inspired by Supervisor.

Developed as part of the **School 42** curriculum, Taskmaster daemonizes, controls, and monitors background processes to ensure high availability and precise lifecycle management. It is designed to handle unexpected crashes, manage detailed environment contexts, and update configurations on the fly without downtime.

## 🚀 Key Technical Features

This project demonstrates a deep understanding of OS-level programming and concurrent architecture:

* **Zero-Downtime Hot Reloading:** Dynamically parses YAML configuration updates and reconstructs process memory states without interrupting unchanged running services.
* **Advanced UNIX Signal Routing:** Manages process groups (`Setpgid`) to prevent zombie/orphan processes. Routes specific signals (e.g., `SIGTERM`, `SIGINT`, `SIGKILL`) dynamically as defined in the configuration.
* **Concurrency & Thread-Safety:** Built with Go's `goroutines` and `sync.Mutex` to allow fully asynchronous process monitoring while safely handling user inputs via an interactive CLI.
* **Environment & State Isolation:** Injects custom environment variables (`env`) and manages file creation masks (`umask`) on a strictly per-process basis.
* **I/O Redirection:** Safely routes and persists `stdout` and `stderr` streams to dedicated, isolated log files.

## 🛠️ Tech Stack

* **Language:** Go (Golang)
* **Configuration:** YAML
* **OS Target:** UNIX/Linux

## ⚙️ Configuration Example

Taskmaster reads a straightforward `yaml` file to define process behaviors:

```yaml
programs:
  my_worker:
    cmd: "./scripts/worker.sh"
    numprocs: 2
    autostart: true
    autorestart: "unexpected"
    starttime: 5
    stopsignal: "TERM"
    stdout: "./logs/worker.stdout.log"
    stderr: "./logs/worker.stderr.log"
    umask: "022"
    env:
      ENV_VAR: "production"

```

## 💻 Quick Start

1. **Clone the repository:**
```bash
git clone https://github.com/your-username/taskmaster-go.git
cd taskmaster-go

```


2. **Run the daemon & CLI:**
```bash
go run cmd/taskmaster/main.go

```


3. **Interactive Commands:**
Once inside the `taskmaster>` prompt, you can manage your services:
* `status` : View the real-time state, PID, and uptime of all processes.
* `start <process_name>` : Manually boot a stopped process.
* `stop <process_name>` : Gracefully terminate a process using its configured UNIX signal.
* `reload` : Apply configuration changes on the fly.
* `exit` : Safely shutdown all managed processes and exit the program.



## 🧠 Architecture Overview

The core of Taskmaster relies on a central `ProcessManager` acting as the source of truth. It abstracts the standard `os/exec` package to wrap processes in custom Goroutines, evaluating execution success via configurable timers (`starttime`) and recovering from unexpected exit codes based on user-defined restart policies (`autorestart`).