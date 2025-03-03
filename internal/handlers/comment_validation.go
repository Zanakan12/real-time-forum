package handlers

import (
	"db"
	"fmt"
	"middlewares"
	"net/http"
	"strconv"
)

func CommentValidationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session := middlewares.GetCookie(w, r)
	if session.Username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	postID, content := r.FormValue("post_id"), r.FormValue("content")
	convPostID, _ := strconv.Atoi(postID)
	err := db.CommentInsert(session.UserID, convPostID, content)
	if err != nil {
		fmt.Printf("error while creating a comment: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
