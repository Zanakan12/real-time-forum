package handlers

import (
	"db"
	"io"
	"middlewares"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Fonction pour vérifier le type d'image
func isValidImageType(header *multipart.FileHeader) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/bim":  true,
	}

	// Ouvrir le fichier pour lire le type MIME
	file, err := header.Open()
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		return false
	}

	fileType := http.DetectContentType(buf)
	return allowedTypes[fileType]
}

// PostValidationHandler handles the validation and insertion of a post
func PostValidationHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the method is POST
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session := middlewares.GetCookie(w, r)
	if session.Username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Retrieve the form data submitted
	postBody := r.FormValue("body") // User ID from form
	title := makeTitle(postBody)    // Post title --- first three letters of the body

	// Récupérer les catégories sélectionnées
	selectedMoods := r.Form["moods"]
	categories := []int{}
	for _, mood := range selectedMoods {
		category, _ := strconv.Atoi(mood)
		categories = append(categories, category)
	}

	var imagePath string
	var fileSize int

	file, header, err := r.FormFile("image")
	if err == nil && file != nil {
		defer file.Close()

		// Vérifier la taille du fichier
		if header.Size > 20*1024*1024 { // 20 Mo en octets
			errorMsg := url.QueryEscape("Img size > 20mb")
			http.Redirect(w, r, "/?error="+errorMsg, http.StatusSeeOther)
			return
		}

		// Vérifier le type de l'image
		if !isValidImageType(header) {
			errorMsg := url.QueryEscape("Img invalid extension")
			http.Redirect(w, r, "/?error="+errorMsg, http.StatusSeeOther)
			return
		}

		// Créer le répertoire d'uploads si nécessaire
		uploadsDir := "./static/assets/img" // Ajustez le chemin si nécessaire
		if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Créer un nouveau fichier dans le répertoire d'uploads
		dst, err := os.Create(filepath.Join(uploadsDir, header.Filename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copier le fichier téléchargé vers le fichier de destination
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		imagePath = filepath.Join(uploadsDir, header.Filename)

		// Enregistrer l'image dans la base de données
		imageFileInfo, err := os.Stat(dst.Name())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fileSize = int(imageFileInfo.Size())
	}

	// Insérer le post et obtenir l'ID
	postID, err := db.PostInsert(session.UserID, title, postBody, categories)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Gérer l'upload de fichier
	if file != nil {
		err = db.ImageInsert(postID, fileSize, imagePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// Redirect to the homepage on success
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func makeTitle(content string) string {
	maxLength := 20                  // Set a maximum character limit
	words := strings.Fields(content) // Split the content into words by spaces

	if len(words) >= 3 {
		title := strings.Join(words[:3], " ") // Join the first three words with a space
		if len(title) > maxLength {           // Check if title exceeds the max length
			return title[:maxLength] + "..." // Truncate the title to the max length
		}
		return title + "..."
	}

	if len(content) > maxLength { // If content has less than two words but exceeds max length
		return content[:maxLength] // Truncate content to the max length
	}

	return content // Return the original content if it's within the limits
}
