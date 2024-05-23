package memoryDB

import (
	"context"
	"errors"
	"forum/model"
)

type posts struct {
	posts map[uint]post
}

type post struct {
	title          string
	content        string
	CommentsLocked bool
	comment        *comment
}

type comments struct {
	comments map[uint]comment
}

type comment struct {
	content  string
	postID   uint
	parentId uint
}

type Data struct {
	currentId uint
	posts     posts
	comments  comments
}

func NewData(current_id uint) *Data {
	return &Data{currentId: 0}
}

func (d *Data) CreatePost(title string, content string, commentsLocked *bool) (*model.Post, error) {
	if len(title) > 255 || len(content) > 2000 || len(title) < 3 {
		return nil, errors.New("wrong lenght of title or content")
	}
	commentsLock := false
	if commentsLocked != nil {
		commentsLock = *commentsLocked
	}
	d.currentId++
	d.posts.posts[d.currentId] = post{
		title:          title,
		content:        content,
		CommentsLocked: commentsLock,
	}
	return &model.Post{
		ID:             d.currentId,
		Title:          title,
		Content:        content,
		CommentsLocked: commentsLock,
	}, nil

}

func (d *Data) CreateComment(ctx context.Context, postID uint, parentID *uint, content string) (*model.CommentConnection, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Data) LockComments(ctx context.Context, postID uint) (*model.Post, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Data) Posts() ([]*model.Post, error) {
	if d.posts.posts == nil {
		return nil, nil
	}
	res := make([]*model.Post, len(d.posts.posts))
	for id, post := range d.posts.posts {
		res = append(res, &model.Post{
			ID:             id,
			Title:          post.title,
			Content:        post.content,
			CommentsLocked: post.CommentsLocked,
		})
	}
	return res, nil
}

func (d *Data) Post(ctx context.Context, id uint) (*model.Post, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Data) Comments(ctx context.Context, id *uint, first *int, after *string) (*model.CommentConnection, error) {
	//TODO implement me
	panic("implement me")
}
