#!/bin/bash

# Fonction déclenchée à la réception d'un SIGTERM
on_term() {
    echo "[$(date +%T)] Worker: J'ai reçu un SIGTERM ! Je m'arrête proprement."
    exit 0
}

# Fonction déclenchée à la réception d'un SIGINT
on_int() {
    echo "[$(date +%T)] Worker: Oh ! Quelqu'un a fait un Ctrl+C (SIGINT). Je coupe tout !"
    exit 0
}

# On dit au script d'associer les fonctions aux signaux correspondants
trap on_term SIGTERM
trap on_int SIGINT

echo "Worker démarré avec le PID $$..."
echo "Ma variable d'environnement custom est : $ENV_VAR"
# Boucle infinie pour maintenir le script en vie
while true; do
    sleep 1
done