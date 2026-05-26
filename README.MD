
# 💀 THE DOOM FORUM

Forum web complet développé en Go, seul, dans le cadre d'un projet scolaire Ynov.

## 👤 Auteur

- **Ulysse** — projet solo

## 🛠️ Stack technique

- **Langage** : Go 1.25.5
- **Base de données** : SQLite
- **Containerisation** : Docker
- **Frontend** : HTML/CSS pur (aucun framework)
- **Auth** : Sessions cookie + OAuth Google

## ✅ Fonctionnalités

### Obligatoires
- Inscription / Connexion / Déconnexion
- Création de posts avec catégories
- Commentaires sur les posts
- Système de likes et dislikes (posts & commentaires)
- Filtrage des posts (par catégorie, mes posts, posts likés)
- Gestion des erreurs HTTP (404, 500...)
- Containerisation Docker

### Bonus implémentés
- 🔐 **OAuth Google** — connexion via compte Google
- 🖼️ **Upload d'images** — dans les posts ET les commentaires (JPEG, PNG, GIF, max 20 Mo)
- 🛡️ **Système de modération complet** :
  - 4 rôles : Visiteur, Utilisateur, Modérateur, Administrateur
  - Les modérateurs peuvent supprimer posts et commentaires
  - Les admins peuvent promouvoir/rétrograder des modérateurs
  - Système de mute/démute des utilisateurs
  - Panel d'administration dédié
  - Signalement de contenu

## 🚀 Lancement avec Docker

```bash
docker-compose up --build
```

L'application sera disponible sur **http://localhost:8080**

## 🚀 Lancement sans Docker

### Prérequis
- Go 1.25.5+
- Git

### Installation

```bash
git clone <url-du-repo>
cd THE-DOOM-FORUM
```

Crée un fichier `.env` à la racine :

```env
CLIENT_ID=ton_google_client_id
CLIENT_SECRET=ton_google_client_secret
```

Lance le serveur :

```bash
cd Projet
go run cmd/main.go
```

L'application sera disponible sur **http://localhost:8080**

## 🔑 Compte admin par défaut

| Email | Mot de passe |
|-------|-------------|
| admin@forum.com | admin123 |


## 📁 Structure du projet

```
THE-DOOM-FORUM/
├── .env                  # Variables d'environnement (non versionné)
├── README.md
└── Projet/
    ├── cmd/main.go       # Point d'entrée
    ├── internal/
    │   ├── database/     # Init BDD + migrations
    │   ├── handlers/     # Logique HTTP
    │   ├── middleware/   # Auth middleware
    │   ├── models/       # Structs Go
    │   └── utils/        # Sessions
    └── web/
        ├── static/       # CSS + uploads
        └── templates/    # HTML templates
```

## 📦 Dépendances Go

```bash
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid
go get github.com/joho/godotenv
```
