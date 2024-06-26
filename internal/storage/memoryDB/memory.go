package memoryDB

import (
	"errors"
	"forum/internal/model"
	"strconv"
	"sync"
)

type MemoryDB struct {
	posts            map[uint]model.Post
	comments         map[uint]*model.Comment
	currentPostId    uint
	currentCommentId uint
	mu               sync.RWMutex
}

func (m *MemoryDB) Stop() error {
	return nil
}

func (m *MemoryDB) CheckPost(postId uint) bool {
	_, ok := m.posts[postId]
	return ok
}

func New() *MemoryDB {
	return &MemoryDB{currentPostId: 1, currentCommentId: 1, posts: make(map[uint]model.Post), comments: make(map[uint]*model.Comment)}
}

func (m *MemoryDB) CreatePost(title string, content string, commentsLocked *bool) (*model.Post, error) {
	if commentsLocked == nil {
		commentsLocked = new(bool)
		*commentsLocked = false
	}
	post := model.Post{Title: title, Content: content, CommentsLocked: *commentsLocked, HasComments: false}
	m.mu.Lock()
	defer m.mu.Unlock()
	post.ID = m.currentPostId
	m.posts[m.currentPostId] = post
	_, ok := m.posts[m.currentPostId]
	if ok {
		m.currentPostId++
		return &post, nil
	}
	return nil, errors.New("can't add post")
}

func (m *MemoryDB) CreateComment(postID uint, parentID *uint, content string) (*model.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	post, ok := m.posts[postID]
	if !ok {
		return nil, errors.New("post not found")
	}
	if post.CommentsLocked {
		return nil, errors.New("comments locked")
	}
	comment := model.Comment{
		ID:       m.currentCommentId,
		PostID:   postID,
		ParentID: parentID,
		Content:  content,
	}
	if parentID == nil {
		post.Comments = append(post.Comments, &comment)
	} else {
		parentComment, ok := m.comments[*parentID]
		if !ok {
			return nil, errors.New("parent comment not found")
		}
		parentComment.Replies = append(parentComment.Replies, &comment)
		m.comments[*parentID] = parentComment
	}
	post.HasComments = true
	m.posts[postID] = post
	m.comments[comment.ID] = &comment
	m.currentCommentId++
	return &comment, nil
}

func (m *MemoryDB) LockComments(postID uint) (*model.Post, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	post, ok := m.posts[postID]
	if !ok {
		return nil, errors.New("post not found")
	}
	post.CommentsLocked = true
	m.posts[postID] = post
	return &post, nil
}

func (m *MemoryDB) Post(id uint, limit *int) (*model.Post, error) {
	if limit == nil {
		limit = new(int)
		*limit = 10
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	post, ok := m.posts[id]
	if !ok {
		return nil, errors.New("post not found")
	}
	if len(post.Comments) > *limit {
		post.Comments = post.Comments[0:*limit]
	}
	return &post, nil
}

func (m *MemoryDB) Posts() ([]*model.Post, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.posts) == 0 {
		return nil, nil
	}
	var res []*model.Post
	for i := uint(1); i < m.currentPostId; i++ {
		post, ok := m.posts[i]
		if ok {
			res = append(res, &post)
		}
	}
	return res, nil
}

func (m *MemoryDB) Comments(postID uint, first *int, after *string) (*model.CommentConnection, error) {
	post, ok := m.posts[postID]
	if len(post.Comments) == 0 {
		return nil, nil
	}
	if !ok {
		return nil, errors.New("post not found")
	}
	if first == nil {
		first = new(int)
		*first = 10
	}
	res := model.CommentConnection{PageInfo: &model.PageInfo{}}
	cursor := new(string)
	if after == nil {
		if len(post.Comments) <= *first {
			res.Comments = post.Comments
			res.PageInfo.HasNextPage = false
		} else {
			res.Comments = post.Comments[0:*first]
			res.PageInfo.HasNextPage = true
			*cursor = strconv.Itoa(*first)
		}

	} else {
		afterInt, err := strconv.Atoi(*after)
		if err != nil || afterInt < 0 || afterInt >= len(post.Comments) {
			return nil, errors.New("invalid cursor")
		}
		if len(post.Comments) <= afterInt+*first {
			res.Comments = post.Comments[afterInt:]
			res.PageInfo.HasNextPage = false
		} else {
			res.Comments = post.Comments[afterInt : afterInt+*first]
			res.PageInfo.HasNextPage = true
			*cursor = strconv.Itoa(afterInt + *first)
		}

	}
	if *cursor == "" {
		cursor = nil
	}
	res.PageInfo.EndCursor = cursor
	return &res, nil

}
