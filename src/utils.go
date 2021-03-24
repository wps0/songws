package main

import (
	"encoding/json"
	"strconv"
	"sync"
)

type DQueue struct {
	data []StatusTrack
	mu   sync.RWMutex
}

func (s *DQueue) Push(val StatusTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy(s.data[1:], s.data[:len(s.data)])
	s.data[0] = val
}

func (s *DQueue) Back() *StatusTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.data[len(s.data)-1]
}

func (s *DQueue) Front() *StatusTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.data[0]
}

func (s *DQueue) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *DQueue) Empty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.data {
		if v.Song != "" {
			return false
		}
	}
	return true
}

func (s *DQueue) To_Json() string {
	update := StatusUpdate{
		Status: 1,
		Data:   make([]StatusTrack, 3),
	}
	s.mu.RLock()
	for i, val := range s.data {
		update.Data[i] = val
	}
	s.mu.RUnlock()
	js, _ := json.Marshal(update)
	return string(js)
}

func track_to_status_track(t Track) StatusTrack {
	date, _ := strconv.Atoi(t.Date.Uts)
	return StatusTrack{
		Uid:            t.Mbid,
		Artist:         t.Artist.Text,
		Song:           t.Name,
		Streaming:      len(t.Attr.Nowplaying) > 0 && t.Attr.Nowplaying[0] == 't',
		StartTimestamp: date,
	}
}
