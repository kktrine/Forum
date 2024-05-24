package data

import "forum/model"

type Storage interface {
	CreatePost(title string, content string, commentsLocked *bool) (*model.Post, error)
	CreateComment(postID uint, parentIDI *uint, parentIDS *string, content string) (*model.Comment, error)
	LockComments(postID uint) (*model.Post, error)
	Post(id uint) (*model.Post, error)
	Posts() ([]*model.Post, error)
	Comments(id *uint, first *int, after *string) (*model.CommentConnection, error)
}
