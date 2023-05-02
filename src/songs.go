package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
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

var dq DQueue = DQueue{
	data: make([]StatusTrack, 3, 3),
}

func fetcher(srv *Server, cfg *Configuration) {
	client := &http.Client{}
	playing_status_update_sent := false
	log.Printf("[Songs fetcher] Starting up...\n")

	for {
		time.Sleep(time.Duration(cfg.RequestInterval) * time.Second)

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

		if rt_resp.Error != 0 {
			log.Printf("[Songs fetcher] Last.fm api error: %d\n", rt_resp.Error)
			continue
		}
		if err != nil {
			log.Printf("[Songs fetcher] JSON Parsing error: %s\n", err)
			continue
		}

		tracks := make([]StatusTrack, len(rt_resp.Recenttracks.Track))
		for i, t := range rt_resp.Recenttracks.Track {
			tracks[i] = track_to_status_track(t)
		}

		sort.Slice(tracks, func(i, j int) bool {
			if tracks[i].StartTimestamp == 0 {
				return true
			} else if tracks[j].StartTimestamp == 0 {
				return false
			}
			return tracks[i].StartTimestamp > tracks[j].StartTimestamp
		})

		// for i, t := range tracks {
		// 	log.Printf("%d: (%d; %v) %s %s (%s)\n", i, t.StartTimestamp, t.Streaming, t.Artist, t.Song, t.Uid)
		// }

		update := StatusUpdate{
			Status: -1,
			Data:   make([]StatusTrack, 0, 3),
		}
		
		if dq.Empty() {
			update.Status = 1
			for i := 0; i < len(tracks); i++ {
				if i >= 3 {
					break
				}
				if tracks[i].StartTimestamp == 0 {
					tracks[i].StartTimestamp = int(time.Now().Unix())
				}
				update.Data = append(update.Data, tracks[i])
			}
			for i := len(tracks) - 1; i >= 0; i-- {
				dq.Push(tracks[i])
			}
		} else {
			if dq.Front().Uid != tracks[0].Uid && !tracks[0].Streaming && !playing_status_update_sent || dq.Front().Uid == tracks[0].Uid && !tracks[0].Streaming && dq.Front().Streaming {
				update.Status = 0
				dq.Front().Streaming = false
				playing_status_update_sent = true
			} else if dq.Front().Hash() != tracks[0].Hash() {
				if tracks[0].StartTimestamp == 0 {
					tracks[0].StartTimestamp = int(time.Now().Unix())
				}
				if dq.Front().StartTimestamp >= tracks[0].StartTimestamp {
					continue
				}
				dq.mu.Lock()
				if dq.data[0].Streaming {
					dq.data[0].Streaming = false
				}
				dq.mu.Unlock()
				dq.Push(tracks[0])

				update.Status = 1
				update.Data = append(update.Data, tracks[0])
				playing_status_update_sent = false
			} else {
				if !dq.Front().Streaming && tracks[0].Streaming {
					update.Status = 0
					playing_status_update_sent = false
					dq.Front().Streaming = true
				}
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
