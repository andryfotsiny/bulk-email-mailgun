#!/bin/bash

# Script pour configurer SSL avec Let's Encrypt
# Usage: ./setup-ssl.sh votre-domaine.com votre-email@example.com

set -e

DOMAIN=$1
EMAIL=$2

if [ -z "$DOMAIN" ] || [ -z "$EMAIL" ]; then
    echo "❌ Usage: ./setup-ssl.sh votre-domaine.com votre-email@example.com"
    exit 1
fi

echo "Configuration SSL pour $DOMAIN"
echo "Email: $EMAIL"
echo ""

# Créer les répertoires nécessaires
echo "Création des répertoires..."
mkdir -p certbot/conf
mkdir -p certbot/www
mkdir -p nginx/logs

# Vérifier que le domaine pointe vers ce serveur
echo "Vérification DNS..."
DOMAIN_IP=$(dig +short $DOMAIN | tail -n1)
SERVER_IP=$(curl -s ifconfig.me)

if [ "$DOMAIN_IP" != "$SERVER_IP" ]; then
    echo "  ATTENTION: Le domaine $DOMAIN ne pointe pas vers ce serveur"
    echo "   IP du domaine: $DOMAIN_IP"
    echo "   IP du serveur: $SERVER_IP"
    read -p "Continuer quand même? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Démarrer Nginx en mode HTTP uniquement pour la validation
echo "Démarrage de Nginx (HTTP)..."
docker-compose up -d nginx

# Attendre que Nginx démarre
echo "⏳ Attente du démarrage de Nginx..."
sleep 5

# Obtenir le certificat SSL
echo "Obtention du certificat SSL..."
docker-compose run --rm certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    -d $DOMAIN \
    -d www.$DOMAIN

if [ $? -eq 0 ]; then
    echo "Certificat SSL obtenu avec succès!"

    # Remplacer la configuration Nginx par la version SSL
    echo "Activation de la configuration SSL..."
    cp nginx/nginx-ssl.conf nginx/nginx.conf

    # Remplacer le domaine dans la config
    sed -i "s/votre-domaine.com/$DOMAIN/g" nginx/nginx.conf

    # Redémarrer avec SSL
    echo " Redémarrage avec SSL..."
    docker-compose -f docker-compose-ssl.yml down
    docker-compose -f docker-compose-ssl.yml up -d

    echo ""
    echo " Configuration SSL terminée!"
    echo " Votre site est maintenant accessible sur:"
    echo "   https://$DOMAIN"
    echo "   https://www.$DOMAIN"
    echo ""
    echo "Le renouvellement automatique est configuré (tous les 12h)"
else
    echo "❌ Erreur lors de l'obtention du certificat SSL"
    echo "Vérifiez que:"
    echo "  - Le domaine $DOMAIN pointe bien vers ce serveur"
    echo "  - Les ports 80 et 443 sont ouverts"
    echo "  - Aucun autre service n'utilise ces ports"
    exit 1
fi