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
	var comments []model.Comment
	query := d.db.Order("id ASC").Model(&comments)
	var ParentId uint
	if id != nil {
		ParentId = *id
	} else if after != nil {
		var firstComment model.Comment
		decodedCursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return nil, errors.New("can't decode \"after\" field: " + err.Error())
		}
		firstIdUint64, err := strconv.ParseUint(string(decodedCursor), 10, 64)
		if err != nil {
			return nil, errors.New("can't parse cursor: " + err.Error())
		}
		firstId := uint(firstIdUint64)
		err = d.db.Where("id = ?", firstId).First(&firstComment).Error
		if err != nil {
			return nil, errors.New("can't find comment: " + err.Error())
		}
		if firstComment.ParentIDI != nil {
			ParentId = *firstComment.ParentIDI
		} else {
			ParentId = firstComment.ID
		}
		query = query.Where("id >= ?", firstComment.ID)
	} else {
		return nil, errors.New("one of \"id\", \"after\" must be provided")
	}
	query = query.Where("parent_id_i = ?", ParentId)
	if first != nil {
		query = query.Limit(*first + 2)
	} else {
		query = query.Limit(12)
	}
	err := query.Find(&comments).Error
	if err != nil {
		return nil, errors.New("can't find comments: " + err.Error())
	}
	res := model.CommentConnection{}
	res.PageInfo = &model.PageInfo{HasNextPage: false}
	if len(comments) > *first+1 {
		res.PageInfo.HasNextPage = true
		comments = comments[0 : *first+1]
	}
	commentEdgeArr := make([]*model.CommentEdge, len(comments))
	for i, comment := range comments {
		commentEdgeArr[i] = &model.CommentEdge{}
		commentEdgeArr[i].Node = &comment
		commentEdgeArr[i].Cursor = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", comment.ID)))
	}
	res.Comments = commentEdgeArr
	res.PageInfo.EndCursor = &commentEdgeArr[len(commentEdgeArr)-1].Cursor
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
