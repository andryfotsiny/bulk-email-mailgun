.PHONY: help install start stop restart logs clean

help: ## Affiche l'aide
	@echo "Commandes disponibles:"
	@echo "  make install    - Installation complète"
	@echo "  make start      - Démarrer l'application"
	@echo "  make stop       - Arrêter l'application"
	@echo "  make restart    - Redémarrer"
	@echo "  make logs       - Voir les logs"
	@echo "  make clean      - Nettoyer tout"
	@echo "  make backup     - Sauvegarder la DB"
	@echo "  make ssl        - Configurer SSL (DOMAIN=... EMAIL=...)"

install: ## Installation complète
	@echo "📦 Installation..."
	@mkdir -p nginx/logs certbot/conf certbot/www backups
	@chmod +x setup-ssl.sh
	docker-compose build
	docker-compose up -d
	@echo " Installation terminée!"
	@echo " Accès: http://localhost/8080"

start: ## Démarrer
	@echo " Démarrage..."
	docker-compose up -d
	@echo "Application démarrée sur http://localhost"

stop: ## Arrêter
	@echo " Arrêt..."
	docker-compose down

restart: stop start ## Redémarrer

logs: ## Voir les logs
	docker-compose logs -f

logs-app: ## Logs de l'app Go uniquement
	docker-compose logs -f bulk-email-app

logs-nginx: ## Logs de Nginx uniquement
	docker-compose logs -f nginx

status: ## Statut des containers
	docker-compose ps

backup: ## Sauvegarder la base de données
	@echo " Sauvegarde..."
	@mkdir -p backups
	@cp emails.db backups/emails-$(shell date +%Y%m%d-%H%M%S).db
	@echo " Sauvegarde créée dans backups/"

clean: ## Nettoyer tout
	@echo " Nettoyage..."
	docker-compose down -v
	docker system prune -f
	@echo " Nettoyage terminé"

rebuild: clean install ## Tout reconstruire

shell: ## Ouvrir un shell dans l'app
	docker-compose exec bulk-email-app /bin/sh

ssl: ## Configurer SSL (Usage: make ssl DOMAIN=example.com EMAIL=you@email.com)
	@if [ -z "$(DOMAIN)" ] || [ -z "$(EMAIL)" ]; then \
		echo "❌ Usage: make ssl DOMAIN=votre-domaine.com EMAIL=vous@email.com"; \
		exit 1; \
	fi
	@echo " Configuration SSL..."
	./setup-ssl.sh $(DOMAIN) $(EMAIL)