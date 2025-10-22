
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

make help - Affiche l'aide

make install - Installation complète

make start - Démarrer l'application

make stop - Arrêter l'application

make restart - Redémarrer l'application

make rebuild - Tout reconstruire (clean + install)

make logs - Voir les logs

make logs-app - Voir les logs de l'app Go uniquement

make logs-nginx - Voir les logs de Nginx uniquement

make status - Statut des containers

make backup - Sauvegarder la base de données

make clean - Nettoyer tout

make shell - Ouvrir un shell dans l'app

make ssl - Configurer SSL (DOMAIN=... EMAIL=...)