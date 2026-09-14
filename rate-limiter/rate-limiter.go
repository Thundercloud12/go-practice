package ratelimiter

import (
	"container/list"
	"sync"
	"time"
)



type clientLog struct{
	mu sync.Mutex
	timestamps  *list.List
}


type RateLimiter struct{
	mapMU sync.Mutex
	clients map[string]*clientLog
	maxRequests int
	window time.Duration
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		mapMU:       sync.Mutex{},
		clients:     make(map[string]*clientLog),
		maxRequests: maxRequests,
		window:      window,
	}
}

func (r *RateLimiter)Allow(clientId string) bool  {
	r.mapMU.Lock()

	mp,ok:=r.clients[clientId]
	if !ok{
		
		mp = &clientLog{
			mu:sync.Mutex{},
			timestamps:list.New(),
		}
		r.clients[clientId]=mp
		
	}

	r.mapMU.Unlock()
	
	mp.mu.Lock()
	tz_map:=mp.timestamps
	
	defer mp.mu.Unlock()
	for e := tz_map.Back(); e != nil; {
		prev := e.Prev()

		if time.Since(e.Value.(time.Time)) > r.window {
			tz_map.Remove(e)
		} else {
			break
		}

		e = prev
	}
	
	if mp.timestamps.Len() < r.maxRequests {
		mp.timestamps.PushFront(time.Now())
		return true		
	}else{
		return false
	}
	
}