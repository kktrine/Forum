package data

import (
	model2 "forum/internal/model"
)

type Storage interface {
	CreatePost(title string, content string, commentsLocked *bool) (*model2.Post, error)
	CreateComment(postID uint, parentID *uint, content string) (*model2.Comment, error)
	LockComments(postID uint) (*model2.Post, error)
	Post(id uint, limit *int) (*model2.Post, error)
	Posts() ([]*model2.Post, error)
	Comments(postID uint, first *int, after *string) (*model2.CommentConnection, error)
	CheckPost(postId uint) bool
}
