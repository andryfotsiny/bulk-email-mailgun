# Build stage
FROM golang:1.25-alpine AS builder

# Installation des dépendances système nécessaires pour SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./

# Télécharger les dépendances
RUN go mod download

# Copier tout le code source
COPY . .

# Build de l'application avec des tags pour Alpine/musl
RUN CGO_ENABLED=1 GOOS=linux go build -tags "sqlite_omit_load_extension" -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

# Installation de SQLite et ca-certificates (pour HTTPS)
RUN apk --no-cache add ca-certificates sqlite-libs

# Créer un utilisateur non-root pour la sécurité
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/data ./data

# Créer le répertoire pour la DB et donner les permissions AVANT de changer d'utilisateur
RUN mkdir -p /app/data && \
    touch /app/emails.db && \
    chown -R appuser:appgroup /app && \
    chmod -R 755 /app && \
    chmod 644 /app/emails.db

USER appuser

EXPOSE 8080

# Healthcheck pour vérifier que l'application fonctionne
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/login || exit 1

# Commande de démarrage
CMD ["./main"]