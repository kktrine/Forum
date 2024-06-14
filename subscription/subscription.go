package subscription

import (
	"forum/model"
	"sync"
)

var (
	CommentSubscribers = make(map[uint][]chan *model.Comment)
	Mu                 sync.Mutex
)
