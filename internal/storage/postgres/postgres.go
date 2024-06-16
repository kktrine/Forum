package postgres

import (
	"errors"
	"fmt"
	model2 "forum/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strconv"
)

type Postgres struct {
	db *gorm.DB
}

func New(cfg string) *Postgres {
	db, err := gorm.Open(postgres.Open(cfg), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("couldn't connect to database: " + err.Error())
	}
	err = db.AutoMigrate(&model2.Post{}, &model2.Comment{})
	if err != nil {
		panic("failed to migrate tables: " + err.Error())
	}
	return &Postgres{db: db}
}

func (d Postgres) Stop() error {
	val, err := d.db.DB()
	if err != nil {
		return errors.New("failed to get database error: " + err.Error())
	}
	return val.Close()
}

func (d Postgres) CheckPost(postId uint) bool {
	var count int64
	err := d.db.Model(&model2.Post{}).Where("id = ?", postId).Count(&count).Error
	if err != nil {
		fmt.Println("Ошибка выполнения запроса:", err)
		return false
	}
	return count > 0
}

func (d Postgres) CreatePost(title string, content string, commentsLocked *bool) (*model2.Post, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	newPost := &model2.Post{
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
		return nil, errors.New("can't insert storage: " + err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("can't commit transaction: " + err.Error())
	}
	return newPost, nil
}

func (d Postgres) CreateComment(postID uint, parentIDI *uint, content string) (*model2.Comment, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	var post model2.Post
	if err := d.db.First(&post, postID).Error; err != nil {
		return nil, errors.New("can't find post: " + err.Error())
	}
	if parentIDI != nil {
		var comment model2.Comment
		err := d.db.First(&comment, parentIDI).Error
		if err != nil {
			return nil, errors.New("can't find comments: " + err.Error())
		}
	}
	if post.CommentsLocked {
		return nil, errors.New("can't comment this post")
	}
	newComment := model2.Comment{
		PostID:  postID,
		Content: content,
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

func (d Postgres) LockComments(postID uint) (*model2.Post, error) {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("can't start transaction; error: " + tx.Error.Error())
	}
	var post model2.Post
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

func (d Postgres) commentProcess(comments []model2.Comment, limit int) ([]*model2.Comment, bool) {
	var res []*model2.Comment
	hasNextPage := false
	commentsMap := make(map[uint]*model2.Comment)
	for _, comment := range comments {
		if comment.ParentID == nil {
			if len(res) < limit {
				res = append(res, &comment)
			} else {
				hasNextPage = true
			}
		} else if value, ok := commentsMap[*comment.ParentID]; ok {
			value.Replies = append(value.Replies, &comment)
		}
		commentsMap[comment.ID] = &comment
	}
	return res, hasNextPage
}

func (d Postgres) Post(id uint, limit *int) (*model2.Post, error) {
	var post model2.Post
	if err := d.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	if limit == nil || *limit <= 0 {
		limit = new(int)
		*limit = 10
	}
	var comments []model2.Comment
	err := d.db.Where("post_id = ?", id).Order("id ASC").Find(&comments).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &post, nil
		}
		return nil, err
	}
	post.Comments, _ = d.commentProcess(comments, *limit)
	return &post, nil

}

func (d Postgres) Posts() ([]*model2.Post, error) {
	posts := make([]*model2.Post, 0)
	res := d.db.Model(&model2.Post{}).Order("id ASC").Find(&posts)
	if res.Error != nil {
		return nil, res.Error
	}
	return posts, nil
}

func (d Postgres) Comments(postID uint, first *int, after *string) (*model2.CommentConnection, error) {
	if first == nil {
		first = new(int)
		*first = 10
	} else if *first <= 0 {
		return nil, errors.New("invalid first field")
	}
	if !d.CheckPost(postID) {
		return nil, errors.New("post not found")
	}
	var afterUint uint = 0
	if after != nil {
		afterTmp, err := strconv.ParseUint(*after, 10, 64)
		if err != nil {
			return nil, err
		}
		afterUint = uint(afterTmp)
	}
	var comments []model2.Comment
	err := d.db.Where("post_id = ? and id > ?", postID, afterUint).Order("id").Limit(*first + 1).Find(&comments).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	res := &model2.CommentConnection{
		PageInfo: &model2.PageInfo{},
	}
	res.Comments, res.PageInfo.HasNextPage = d.commentProcess(comments, *first)
	if res.PageInfo.HasNextPage {
		cursor := strconv.FormatUint(uint64(res.Comments[len(res.Comments)-1].ID), 10)
		res.PageInfo.EndCursor = &cursor
	}
	return res, err

}
