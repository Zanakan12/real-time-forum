package handlers

import "db"

func countLikesDislikes(posts []db.Post) []db.Post {
	for i, post := range posts {
		for _, likesdislikes := range post.LikesDislikes {
			if likesdislikes.IsLike {
				posts[i].LikesCount++
			} else {
				posts[i].DislikesCount++
			}
		}
		for j, comment := range post.Comments {
			for _, likedislike := range comment.LikesDislikes {
				if likedislike.IsLike {
					posts[i].Comments[j].LikesCount++
				} else {
					posts[i].Comments[j].DislikesCount++
				}
			}
		}
	}
	return posts
}

func countLikesDislikesComments(comments []db.Comment) []db.Comment {
	for j, comment := range comments {
		for _, likedislike := range comment.LikesDislikes {
			if likedislike.IsLike {
				comments[j].LikesCount++
			} else {
				comments[j].DislikesCount++
			}
		}
	}
	return comments
}
