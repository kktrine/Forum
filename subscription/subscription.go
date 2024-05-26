package subscription

import (
	"forum/model"
	"sync"
)

var (
	СommentSubscribers = make(map[uint][]chan *model.Comment)
	Mu                 sync.Mutex
)
