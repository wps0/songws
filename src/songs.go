package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const ROOT_API_URL = "http://ws.audioscrobbler.com/2.0/"

type Track struct {
	Attr struct {
		Nowplaying string `json:"nowplaying"`
	} `json:"@attr"`
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
	Error        int `json:"error"`
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
		update.Data[i] = track_to_status_track(val)
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

		log.Printf("[Songs fetcher] Making last.fm api request...\n")

		req, err := http.NewRequest("GET", ROOT_API_URL+"?method=user.getrecenttracks&user="+cfg.Username+"&api_key="+cfg.LastfmApiKey+"&format=json&limit=3", nil)
		if err != nil {
			log.Printf("[Songs fetcher] An error occurred during creating new request. Error: %d\n", err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Songs fetcher] Request to the API cannot be made. Error: %s\n", err)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[Songs fetcher] An error occurred when parsing response. Error: %s\n", err)
			continue
		}

		rt_resp := RecentTracksResponse{}
		err = json.Unmarshal(body, &rt_resp)

		// log.Println(string(body))

		if rt_resp.Error != 0 {
			log.Printf("[Songs fetcher] Last.fm api error: %d\n", rt_resp.Error)
			continue
		}
		if err != nil {
			log.Printf("[Songs fetcher] JSON Parsing error: %s\n", err)
			continue
		}

		update := StatusUpdate{
			Status: -1,
		}
		if dq.data[0].Mbid == "" {
			update.Status = 1
			update.Data = make([]StatusTrack, 3)
			for i := 0; i < 3; i++ {
				track := rt_resp.Recenttracks.Track[i]
				dq.Push(track)
				update.Data[2-i] = track_to_status_track(track)
			}
		} else {

			if rt_resp.Recenttracks.Track[0].Mbid == dq.Front().Mbid && rt_resp.Recenttracks.Track[0].Attr.Nowplaying != dq.Front().Attr.Nowplaying {
				log.Println("2")
				dq.mu.Lock()
				dq.data[0].Attr.Nowplaying = rt_resp.Recenttracks.Track[0].Attr.Nowplaying
				dq.mu.Unlock()
				update.Status = 0
			} else if rt_resp.Recenttracks.Track[0].Mbid != dq.Front().Mbid {
				log.Println("3")
				dq.Push(rt_resp.Recenttracks.Track[0])
				update.Status = 1
				update.Data = []StatusTrack{track_to_status_track(rt_resp.Recenttracks.Track[0])}
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
