package handlers

import (
	"db"
	"fmt"
	"middlewares"
	"net/http"
	"strconv"
)

func LikesDislikesValidationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session := middlewares.GetCookie(w, r)
	if session.Username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	postID, commentID, likeDislike := r.FormValue("post_id"), r.FormValue("comment_id"), r.FormValue("like_dislike")
	intPostID, intCommentID := -1, -1
	if postID == "" {
		intCommentID, _ = strconv.Atoi(commentID)
	} else {
		intPostID, _ = strconv.Atoi(postID)
	}
	isLike := false
	if likeDislike == "like" {
		isLike = true
	}
	err := db.LikesInsert(session.UserID, intPostID, intCommentID, isLike)
	if err != nil {
		fmt.Printf("error while creating a like/dislike: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
