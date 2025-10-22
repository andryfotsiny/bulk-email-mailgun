
## Installation

```bash

# 1. Clean existing dependencies
rm go.sum

# 2. Reinstall dependencies
go mod tidy
go mod download

# 3. Configure environment variables
nano .env

# 4. Run the application
go run main.go

```

# Affiche l'aide
```bash
make help
```
# Installation complète
make install

# Démarrer l'application
make start

# Arrêter l'application
make stop

# Redémarrer l'application
make restart

# Voir tous les logs
make logs

# Logs de l'app Go uniquement
make logs-app

# Logs de Nginx uniquement
make logs-nginx

# Statut des containers
make status

# Sauvegarder la base de données
make backup

# Nettoyer tout
make clean

# Tout reconstruire
make rebuild

# Ouvrir un shell dans l'app
make shell

# Configurer SSL (à exécuter avec votre domaine et email)
make ssl DOMAIN=votre-domaine.com EMAIL=vous@email.com
