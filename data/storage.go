package data

import "forum/model"

type Storage interface {
	CreatePost(title string, content string, commentsLocked *bool) (*model.Post, error)
	CreateComment(postID uint, parentID *uint, content string) (*model.Comment, error)
	LockComments(postID uint) (*model.Post, error)
	Post(id uint, limit *int) (*model.Post, error)
	Posts() ([]*model.Post, error)
	Comments(postID uint, first *int, after *int) (*model.CommentConnection, error)
}
