#!/bin/bash
echo "Démarrage du système Go..."

if [ ! -f ".env" ]; then
    cp .env.example .env
    echo "Configurez .env avec vos infos SMTP"
fi

go mod download
go run main.go
