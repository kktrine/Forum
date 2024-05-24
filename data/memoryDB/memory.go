package memoryDB

import (
	"context"
	"errors"
	"forum/model"
	"strconv"
	"strings"
)

type posts struct {
	posts map[uint]post
}

type post struct {
	content          string
	title            string
	commentsLocked   bool
	hasComments      bool
	comments         map[string]comment
	currentCommentId uint
}

type comment struct {
	content        string
	currentReplyId uint
	replies        map[string]comment
}

type Data struct {
	currentId uint
	posts     posts
}

func NewData() *Data {
	return &Data{currentId: 0, posts: posts{posts: make(map[uint]post)}}
}

func (d *Data) CreatePost(title string, content string, commentsLocked *bool) (*model.Post, error) {
	if len(title) > 255 || len(content) > 2000 || len(title) < 3 {
		return nil, errors.New("wrong length of title or content")
	}
	commentsLock := false
	if commentsLocked != nil {
		commentsLock = *commentsLocked
	}
	d.currentId++
	d.posts.posts[d.currentId] = post{
		content:          content,
		title:            title,
		commentsLocked:   commentsLock,
		hasComments:      false,
		comments:         nil,
		currentCommentId: 0,
	}

	return &model.Post{
		ID:             d.currentId,
		Title:          title,
		Content:        content,
		HasComments:    false,
		CommentsLocked: commentsLock,
	}, nil

}

func (d *Data) CreateComment(postID uint, parentID *string, content string) (*model.Comment, error) {
	postFound, ok := d.posts.posts[postID]
	if !ok {
		return nil, errors.New("post not exists")
	}
	if len(content) < 3 {
		return nil, errors.New("wrong length of content")
	}
	if parentID == nil {
		postFound.currentCommentId++
		postFound.comments[strconv.FormatUint(uint64(postFound.currentCommentId), 10)] = comment{
			content:        content,
			currentReplyId: 0,
			replies:        nil,
		}
		d.posts.posts[postID] = postFound
		return model.Comment{}, nil
	}
	num := len(*parentID) - strings.Count(*parentID, "_")
	currId := (*parentID)[0:1]
	currComment := postFound.comments[currId]
	for currId != *parentID {

	}

}

func (d *Data) LockComments(postID uint) (*model.Post, error) {
	post, ok := d.posts.posts[postID]
	if !ok {
		return nil, errors.New("post not found")
	}
	post.commentsLocked = true
	d.posts.posts[postID] = post
	return &model.Post{
		ID:             postID,
		Title:          post.title,
		Content:        post.content,
		HasComments:    post.hasComments,
		CommentsLocked: post.commentsLocked,
	}, nil
}

func (d *Data) Posts() ([]*model.Post, error) {
	if d.posts.posts == nil {
		return nil, nil
	}
	res := make([]*model.Post, 0, len(d.posts.posts))
	for id, post := range d.posts.posts {
		res = append(res, &model.Post{
			ID:             id,
			Title:          post.title,
			Content:        post.content,
			HasComments:    post.hasComments,
			CommentsLocked: post.commentsLocked,
		})
	}
	return res, nil
}

func (d *Data) Post(id uint) (*model.Post, error) {
	postRes, ok := d.posts.posts[id]
	if !ok {
		return nil, errors.New("post not found")
	}
	res := model.Post{
		ID:             id,
		Title:          postRes.title,
		Content:        postRes.content,
		HasComments:    postRes.hasComments,
		CommentsLocked: postRes.commentsLocked,
	}
	for idComm, comment := range postRes.comments {
		res.Comments = append(res.Comments, model.Comment{
			ID:      idComm,
			PostID:  id,
			Content: comment.content,
		})
	}
	return &res, nil

}

func (d *Data) Comments(ctx context.Context, id *uint, first *int, after *string) (*model.CommentConnection, error) {
	//TODO implement me
	panic("implement me")
}
