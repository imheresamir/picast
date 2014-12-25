package picast

import (
	"log"
	"strings"
	//"strconv"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/op/go-libspotify/spotify"
	"net/http"
	"reflect"
	"regexp"
	//"sync"
	//"os"
	"io/ioutil"
	//"os/exec"
	//"image/jpeg"
	//"os"
	//"time"
	"encoding/json"
	"errors"
)

func (media *Media) Init() {
	media.Playlist = make([]PlaylistEntry, 0)
}

func (media *Media) GetPlaylist(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(&struct{ Playlist []PlaylistEntry }{Playlist: media.Playlist})
}

func (media *Media) Play(w rest.ResponseWriter, r *rest.Request) {
	if media.Player != nil && media.Player.StatusCode() != STOPPED {
		media.Player.TogglePause()
	} else {
		media.CurrentIndex = 0
		go media.PlayAll()
	}

	w.WriteJson(&ServerStatus{Server: "Media playing."})
}

func (media *Media) Add(w rest.ResponseWriter, r *rest.Request) {
	object := ServerObject{}

	err := r.DecodeJsonPayload(&object)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	switch {
	case object.Url == "":
		rest.NotFound(w, r)
		return

	}

	if object.Data != nil {
		switch object.Data.(type) {
		case map[string]interface{}:
			login := object.Data.(map[string]interface{})

			SpotifyLogin = spotify.Credentials{
				Username: login["Username"].(string),
				Password: login["Password"].(string),
			}
		default:
			log.Println("Invalid Data field")
			break
		}
	}

	parsedEntry, err := media.parseUrl(object.Url)
	if err != nil {
		w.WriteJson(&ServerStatus{Server: err.Error()})
		return
	}

	media.Playlist = append(media.Playlist, parsedEntry)

	//log.Println("Added:", parsedEntry)
	//log.Println("Current Playlist:", media.Playlist)

	go func() {
		media.MediaAdded <- true
	}()

	w.WriteJson(&ServerStatus{Server: "Media added."})

}

func (media *Media) parseUrl(url string) (PlaylistEntry, error) {
	// Sanitize url string
	url = strings.Trim(url, " \n")

	parsedEntry := PlaylistEntry{}

	switch {
	case strings.Contains(url, "spotify"):
		parsedEntry.Url = "spotify"
		var trackId string

		re := regexp.MustCompile(`https?:\/\/open\.spotify\.com\/(\w+)\/(\w+)|spotify:(\w+):(\w+):?(\w+)?:?(\w+)?`)
		matches := re.FindAllStringSubmatch(url, -1)

		for i := range matches {
			for j := range matches[i] {
				if j != 0 && matches[i][j] != "" {
					parsedEntry.Url += ":"
					parsedEntry.Url += matches[i][j]
					trackId = matches[i][j] // TrackId is the last match
				}
			}
		}

		if parsedEntry.Url == "spotify" {
			// parsedUrl hasn't changed
			errorText := "Invalid Spotify uri."
			log.Println(errorText)
			return parsedEntry, errors.New(errorText)
		}

		resp, err := http.Get("https://api.spotify.com/v1/tracks/" + trackId)
		if err != nil {
			errorText := "Could not get Spotify metadata."
			log.Println(errorText)
			return parsedEntry, errors.New(errorText)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		var f interface{}
		err = json.Unmarshal(body, &f)
		if err != nil {
			errorText := "Error decoding JSON"
			log.Println(errorText)
			return parsedEntry, errors.New(errorText)
		}

		m := f.(map[string]interface{})

		for k, v := range m {
			switch k {
			case "name":
				parsedEntry.Title = v.(string)
			case "artists":
				// ["artists"] is an []interface{}
				for i, u := range v.([]interface{}) {
					if i == 0 {
						parsedEntry.Artist = u.(map[string]interface{})["name"].(string)
					} else {
						parsedEntry.Artist = parsedEntry.Artist + ", " + u.(map[string]interface{})["name"].(string)
					}
				}
			case "album":
				// ["album"] is an interface{}
				parsedEntry.Album = v.(map[string]interface{})["name"].(string)
				parsedEntry.ArtPath = v.(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["url"].(string)
			}
		}

		return parsedEntry, nil

	default:
		outfileChan := make(chan string)

		go YoutubeDl(url, outfileChan)
		parsedEntry.Url = <-outfileChan

		if parsedEntry.Url == "" {
			errorText := "Youtube-dl could not find video link."
			log.Println(errorText)
			return parsedEntry, errors.New(errorText)
		}

		// TODO: Parse youtube with youtube api
		return parsedEntry, nil
	}

}

func (media *Media) PlayAll() {
	for ; media.CurrentIndex < len(media.Playlist); media.CurrentIndex++ {
		parsedUrl := media.Playlist[media.CurrentIndex].Url
		log.Println("Playing", media.Playlist[media.CurrentIndex], "[", media.CurrentIndex, "]")

		switch {
		case strings.Contains(parsedUrl, "spotify"):

			if strings.Contains(parsedUrl, "local") == true {
				log.Println("Skipping local track")
				continue
			}

			if reflect.DeepEqual(SpotifyLogin, spotify.Credentials{}) == true {
				log.Println("Could not log in to Spotify.")
				return
			}

			if media.Player != nil && media.Player.StatusCode() != STOPPED {
				switch media.Player.(type) {
				case *SpotifyPlayer:
					media.Player.(*SpotifyPlayer).Outfile = parsedUrl
				default:
					media.Player.Stop(-1)

					media.Player = &SpotifyPlayer{
						Outfile:     parsedUrl,
						KillSwitch:  make(chan int),
						ChangeTrack: make(chan bool),
						PauseTrack:  make(chan bool),
						ResumeTrack: make(chan bool),
						StopTrack:   make(chan bool),
						//ParsePlaylist: make(chan bool),
						//TrackResults: make(chan []string),
					}
				}
			} else {
				media.Player = &SpotifyPlayer{
					Outfile:     parsedUrl,
					KillSwitch:  make(chan int),
					ChangeTrack: make(chan bool),
					PauseTrack:  make(chan bool),
					ResumeTrack: make(chan bool),
					StopTrack:   make(chan bool),
					//ParsePlaylist: make(chan bool),
					//TrackResults: make(chan []string),
				}
			}

			media.MediaChanged <- true
			go media.Player.Play()

		default:

			if media.Player.StatusCode() != STOPPED {
				media.Player.Stop(-1)
			}

			media.Player = &OmxPlayer{Outfile: parsedUrl, KillSwitch: make(chan int)}

			go media.Player.Play()
		}

		switch media.Player.ReturnCode() {
		case -1:
			break
		case 1:
			continue
		}

	}

	if media.CurrentIndex == len(media.Playlist)-1 && media.Player.StatusCode() != STOPPED {
		media.Player.Stop(-1)
	}

}

func (media *Media) Status(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(media.StatusBuilder())
}

// TODO: Needs cleanup
func (media *Media) StatusBuilder() *ServerStatus {
	status := &ServerStatus{Server: "No media."}

	if media.Player != nil {

		switch media.Player.StatusCode() {
		case 0:
			status.Server = "Media stopped."
		case 1:
			status.Server = "Media loading."
		case 2:
			status.Server = "Media paused."
		case 3:
			status.Server = "Media playing."
		}

	}

	return status
}

func (media *Media) TogglePause(w rest.ResponseWriter, r *rest.Request) {
	if media.Player != nil && media.Player.StatusCode() > 1 {
		media.Player.TogglePause()
	}

	w.WriteJson(media.StatusBuilder())
}

func (media *Media) Stop(w rest.ResponseWriter, r *rest.Request) {
	if media.Player != nil /* && media.Player.StatusCode() > 0 */ {
		media.Player.Stop(-1)
		//media.Player = nil
		//media.Player.ReturnCode()
	}

	//w.WriteJson(media.StatusBuilder())
	w.WriteJson(&ServerStatus{Server: "Media stopped."})
}
