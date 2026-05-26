package handlers

import (
	"encoding/json"
	"fmt"
	"forum/internal/utils"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("CLIENT_ID")
	redirectURI := "http://localhost:8080/auth/google/callback"

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid+email+profile",
		clientID,
		url.QueryEscape(redirectURI),
	)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		h.NotFound(w, r)
		return
	}

	// Échange le code contre un token
	token, err := exchangeGoogleCode(code)
	if err != nil {
		h.InternalError(w, r)
		return
	}

	// Récupère les infos de l'utilisateur
	googleUser, err := getGoogleUser(token)
	if err != nil {
		h.InternalError(w, r)
		return
	}

	// Cherche si l'utilisateur existe déjà
	var userID string
	err = h.DB.QueryRow("SELECT id FROM users WHERE email = ?", googleUser["email"]).Scan(&userID)

	if err != nil {
		// Nouvel utilisateur — on le crée
		userID = uuid.New().String()
		username := googleUser["name"]
		if username == "" {
			username = strings.Split(googleUser["email"], "@")[0]
		}

		// Mot de passe aléatoire (inutilisable, c'est voulu)
		fakePassword, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)

		// Vérifie que le username n'est pas déjà pris
		var count int
		h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
		if count > 0 {
			username = username + "_" + userID[:4]
		}

		_, err = h.DB.Exec(
			"INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)",
			userID, googleUser["email"], username, string(fakePassword),
		)
		if err != nil {
			h.InternalError(w, r)
			return
		}
	}

	// Crée la session
	if err := utils.CreateSession(w, h.DB, userID); err != nil {
		h.InternalError(w, r)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func exchangeGoogleCode(code string) (string, error) {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURI := "http://localhost:8080/auth/google/callback"

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("pas de access_token dans la réponse")
	}
	return token, nil
}

func getGoogleUser(accessToken string) (map[string]string, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	user := map[string]string{
		"email": fmt.Sprintf("%v", result["email"]),
		"name":  fmt.Sprintf("%v", result["name"]),
	}
	return user, nil
}
