package subscription

import (
	"forum/internal/model"
	"sync"
)

var (
	CommentSubscribers = make(map[uint][]chan *model.Comment)
	Mu                 sync.Mutex
)
