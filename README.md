
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
make help           # Voir toutes les commandes
make install        # Installation complète (dev)
make up             # Démarrer (HTTP)
make down           # Arrêter
make logs           # Voir les logs
make status         # Statut des containers
make health         # Vérifier la santé
make backup-db      # Sauvegarder la DB
make ssl-setup      # Configurer SSL (prod)
make prod           # Déployer production

