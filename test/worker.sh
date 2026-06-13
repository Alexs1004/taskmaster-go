#!/bin/bash
# worker.sh : Un script simple qui affiche un compteur toutes les secondes
COUNT=1
echo "Début du worker avec le PID $$"

while true; do
    echo "Le worker tourne... itération $COUNT"
    sleep 1
    ((COUNT++))
done
