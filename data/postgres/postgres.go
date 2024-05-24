package postgres

import (
	"encoding/base64"
	"errors"
	"fmt"
	"forum/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
)

type Db struct {
	db *gorm.DB
}

func (d Db) CreatePost(title string, content string, commentsLocked *bool) (*model.Post, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	newPost := &model.Post{
		Title:       title,
		Content:     content,
		HasComments: false,
		Comments:    nil,
	}
	if commentsLocked != nil {
		newPost.CommentsLocked = *commentsLocked
	}
	if err := tx.Create(&newPost).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("can't insert data: " + err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("can't commit transaction: " + err.Error())
	}
	return newPost, nil
}

func (d Db) CreateComment(postID uint, parentIDI *uint, _ *string, content string) (*model.Comment, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	var post model.Post
	if err := d.db.First(&post, postID).Error; err != nil {
		return nil, errors.New("can't find post: " + err.Error())
	}
	if parentIDI != nil {
		var comment model.Comment
		err := d.db.First(&comment, parentIDI).Error
		if err != nil {
			return nil, errors.New("can't find comments: " + err.Error())
		}
	}
	if post.CommentsLocked {
		return nil, errors.New("can't comment this post")
	}
	newComment := model.Comment{
		PostID:    postID,
		ParentIDI: parentIDI,
		Content:   content,
	}
	err := tx.Create(&newComment).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("can't insert comment: " + err.Error())
	}
	if post.HasComments == false {
		post.HasComments = true
		tx.Save(&post)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("can't commit transaction: " + err.Error())
	}
	return &newComment, nil
}

func (d Db) LockComments(postID uint) (*model.Post, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	var post model.Post
	if err := d.db.First(&post, postID).Error; err != nil {
		return nil, err
	}
	post.CommentsLocked = true
	if err := d.db.Save(&post).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("can't commit transaction: " + err.Error())
	}
	return &post, nil
}

func (d Db) Post(id uint) (*model.Post, error) {
	var post model.Post
	if err := d.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	res := d.db.Where("Post_ID = ? and Parent_id_i is null", id).Find(&post.Comments)
	if res.Error != nil {
		return nil, errors.New("something went wrong: " + res.Error.Error())
	}
	for i, comment := range post.Comments {
		found := d.db.Where("Post_ID = ? and Parent_id_i = ?", id, comment.ID).Find(&comment.Reply)
		if found.RowsAffected > 0 {
			post.Comments[i].Reply = comment.Reply
		}
	}
	return &post, nil
}

func (d Db) Posts() ([]*model.Post, error) {
	posts := make([]*model.Post, 0)
	res := d.db.Order("id").Find(&posts)
	if res.Error != nil {
		return nil, res.Error
	}
	return posts, nil
}

func (d Db) Comments(id *uint, first *int, after *string) (*model.CommentConnection, error) {
	query := d.db.Order("id ASC")
	firstId := 10
	if first != nil {
		firstId = *first
		query = query.Limit(firstId + 1)
	}
	afterId := 0
	if after != nil {
		decode, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, err
		}
		afterId, err = strconv.Atoi(string(decode))
		if err != nil {
			return nil, err
		}
		query = query.Where("id > ?", afterId)
	}
	var comments []model.Comment
	query.Find(&comments)
	if len(comments) == 0 {
		return nil, errors.New("not found")
	}
	var res model.CommentConnection
	if len(comments) > firstId {
		res.PageInfo.HasNextPage = true
		comments = comments[:firstId]
	} else {
		res.PageInfo.HasNextPage = false
	}
	for i, comment := range comments {
		res.Comments[i].Node = &comment
		res.Comments[i].Cursor = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", comment.ID)))
	}
	res.PageInfo.HasNextPage = len(comments) > firstId
	res.PageInfo.EndCursor = &res.Comments[len(res.Comments)-1].Cursor
	return &res, nil
}

func New(cfg string) *Db {
	db, err := gorm.Open(postgres.Open(cfg), &gorm.Config{})
	if err != nil {
		panic("couldn't connect to database: " + err.Error())
	}
	err = db.AutoMigrate(&model.Post{}, &model.Comment{})
	if err != nil {
		panic("failed to migrate tables: " + err.Error())
	}
	db = db.Debug()
	return &Db{db: db}
}

func (d Db) Stop() error {
	val, err := d.db.DB()
	if err != nil {
		return errors.New("failed to get database error: " + err.Error())
	}
	return val.Close()
}
