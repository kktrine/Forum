package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.47

import (
	"context"
	"fmt"
	"forum/model"
)

// Reply is the resolver for the reply field.
func (r *commentResolver) Reply(ctx context.Context, obj *model.Comment) ([]*model.Comment, error) {
	panic(fmt.Errorf("not implemented: Reply - reply"))
}

// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, commentsLocked *bool) (*model.Post, error) {
	return r.Db.CreatePost(title, content, commentsLocked)
}

// CreateComment is the resolver for the createComment field.
func (r *mutationResolver) CreateComment(ctx context.Context, postID uint, parentID *uint, content string) (*model.CommentConnection, error) {
	panic(fmt.Errorf("not implemented: CreateComment - createComment"))
}

// LockComments is the resolver for the lockComments field.
func (r *mutationResolver) LockComments(ctx context.Context, postID uint) (*model.Post, error) {
	return r.Db.LockComments(postID)
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context) ([]*model.Post, error) {
	return r.Db.Posts()
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id uint) (*model.Post, error) {
	return r.Db.Post(id)
}

// Comments is the resolver for the comments field.
func (r *queryResolver) Comments(ctx context.Context, id *uint, first *int, after *string) (*model.CommentConnection, error) {
	panic(fmt.Errorf("not implemented: Comments - comments"))
}

// Comment returns CommentResolver implementation.
func (r *Resolver) Comment() CommentResolver { return &commentResolver{r} }

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type commentResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
