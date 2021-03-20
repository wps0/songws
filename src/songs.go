package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const ROOT_API_URL = "http://ws.audioscrobbler.com/2.0/"

type Track struct {
	Artist struct {
		Mbid string `json:"mbid"`
		Text string `json:"#text"`
	} `json:"artist"`
	Album struct {
		Mbid string `json:"mbid"`
		Text string `json:"#text"`
	} `json:"album"`
	Image []struct {
		Size string `json:"size"`
		Text string `json:"#text"`
	} `json:"image"`
	Streamable string `json:"streamable"`
	Date       struct {
		Uts  string `json:"uts"`
		Text string `json:"#text"`
	} `json:"date"`
	URL  string `json:"url"`
	Name string `json:"name"`
	Mbid string `json:"mbid"`
}

type RecentTracksResponse struct {
	Recenttracks struct {
		Attr struct {
			Page       string `json:"page"`
			Total      string `json:"total"`
			User       string `json:"user"`
			Perpage    string `json:"perPage"`
			Totalpages string `json:"totalPages"`
		} `json:"@attr"`
		Track []Track `json:"track"`
	} `json:"recenttracks"`
}

type DQueue struct {
	data []Track
	mu   sync.RWMutex
}

func (s *DQueue) Push(val Track) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy(s.data[1:], s.data[:len(s.data)])
	s.data[0] = val
}

func (s *DQueue) Back() Track {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[len(s.data)-1]
}

func (s *DQueue) Front() Track {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[len(s.data)-1]
}

func (s *DQueue) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *DQueue) To_Json() string {
	update := StatusUpdate{
		Status: 1,
		Data:   make([]StatusTrack, 3),
	}
	s.mu.RLock()
	for i, val := range s.data {
		now_streaming, _ := strconv.Atoi(val.Streamable)
		update.Data[i] = StatusTrack{
			Artist:    val.Artist.Text,
			Song:      val.Name,
			Streaming: now_streaming > 0,
		}
	}
	s.mu.RUnlock()
	js, _ := json.Marshal(update)
	return string(js)
}

var dq DQueue = DQueue{
	data: make([]Track, 3),
}

func fetcher(srv *Server, cfg *Configuration) {
	client := &http.Client{}

	for {
		time.Sleep(time.Duration(cfg.RequestInterval) * time.Second)

		log.Printf("Making last.fm api request...\n")

		req, err := http.NewRequest("GET", ROOT_API_URL+"?method=user.getrecenttracks&user="+cfg.Username+"&api_key="+cfg.LastfmApiKey+"&format=json&limit=3", nil)
		if err != nil {
			log.Printf("An error occurred during creating new request. Error: %d\n", err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Request to the API cannot be made. Error: %s\n", err)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("An error occurred when parsing response. Error: %s\n", err)
			continue
		}

		rt_resp := RecentTracksResponse{}
		json.Unmarshal(body, &rt_resp)

		update := StatusUpdate{
			Status: -1,
		}
		if dq.data[0].Mbid == "" {
			log.Println("1")
			update.Status = 1
			update.Data = make([]StatusTrack, 3)
			for i, val := range rt_resp.Recenttracks.Track {
				dq.Push(val)
				now_streaming, _ := strconv.Atoi(rt_resp.Recenttracks.Track[0].Streamable)
				update.Data[i] = StatusTrack{
					Artist:    val.Artist.Text,
					Song:      val.Name,
					Streaming: now_streaming > 0,
				}
			}
		} else {

			if rt_resp.Recenttracks.Track[0].Mbid == dq.Front().Mbid && rt_resp.Recenttracks.Track[0].Streamable != dq.Front().Streamable {
				dq.mu.Lock()
				log.Println("2")
				dq.data[0].Streamable = rt_resp.Recenttracks.Track[0].Streamable
				update.Status = 0
				dq.mu.Unlock()
			} else if rt_resp.Recenttracks.Track[0].Mbid != dq.Front().Mbid {
				log.Println("3")
				update.Status = 1
				now_streaming, _ := strconv.Atoi(rt_resp.Recenttracks.Track[0].Streamable)
				update.Data = []StatusTrack{{
					Artist:    rt_resp.Recenttracks.Track[0].Artist.Text,
					Song:      rt_resp.Recenttracks.Track[0].Name,
					Streaming: now_streaming > 0,
				}}
			}
		}
		js, err := json.Marshal(update)

		if err != nil {
			log.Printf("Cannot encode message to json. Error: %s\n", err)
			continue
		}
		if update.Status != -1 {
			srv.broadcast <- string(js)
		}
	}
}
