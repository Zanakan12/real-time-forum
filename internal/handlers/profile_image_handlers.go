package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Fonction qui sauvegarde l'image sous le nom "profileimage"
func saveImageToUserFolder(username string, fileHeader *multipart.FileHeader, file multipart.File) (string, error) {
	// Créer le dossier de l'utilisateur
	userDir := filepath.Join("static", "assets", "img", username)
	err := os.MkdirAll(userDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Erreur lors de la création du dossier : %v", err)
	}

	// Récupérer l'extension du fichier original
	ext := filepath.Ext(fileHeader.Filename) // Exemple : ".jpg" ou ".png"
	if ext == "" {
		ext = ".png" // Sécurité : Si pas d'extension, on met ".png"
	}

	// Nom du fichier toujours "profileimage.extension"
	filePath := filepath.Join(userDir, "profileimage"+ext)

	// Création du fichier destination
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("Erreur lors de la création du fichier : %v", err)
	}
	defer dst.Close()

	// Copier le fichier
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("Erreur lors de l'enregistrement de l'image : %v", err)
	}

	return filePath, nil
}

// Gestionnaire de l'upload
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer le nom d'utilisateur
	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Erreur : Nom d'utilisateur manquant", http.StatusBadRequest)
		return
	}

	// Récupérer le fichier
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Erreur : Aucune image fournie", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Sauvegarde avec nom "profileimage"
	filePath, err := saveImageToUserFolder(username, fileHeader, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse
	fmt.Fprintf(w, "Image enregistrée : %s", filePath)
}
