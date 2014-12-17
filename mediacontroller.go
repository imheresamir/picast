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
	//"io/ioutil"
	//"os/exec"
	//"image/jpeg"
	//"os"
	//"time"
)

func (media *Media) Init() {
	media.Playlist = make([]string, 0)
}

func (media *Media) Play(w rest.ResponseWriter, r *rest.Request) {
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
			//entry.Data = nil

			SpotifyLogin = spotify.Credentials{
				Username: login["Username"].(string),
				Password: login["Password"].(string),
			}
		default:
			log.Println("Invalid Data field")
			break
		}
	}

	// Sanitize url string
	object.Url = strings.Trim(object.Url, " \n")
	log.Println(object)

	media.Playlist = append(media.Playlist, object.Url)
	media.CurrentIndex = len(media.Playlist) - 1

	go media.PlayAll(media.CurrentIndex)

	w.WriteJson(&ServerStatus{Server: "Media playing."})

}

func (media *Media) PlayAll(currentIndex int) {
	for index := currentIndex; index < len(media.Playlist); index++ {
		url := media.Playlist[index]

		switch {
		case strings.Contains(url, "spotify"):
			spotifyUri := "spotify"

			re := regexp.MustCompile(`https?:\/\/open\.spotify\.com\/(\w+)\/(\w+)|spotify:(\w+):(\w+):?(\w+)?:?(\w+)?`)
			matches := re.FindAllStringSubmatch(url, -1)

			for i := range matches {
				for j := range matches[i] {
					if j != 0 && matches[i][j] != "" {
						spotifyUri += ":"
						spotifyUri += matches[i][j]
					}
				}
			}

			if spotifyUri == "spotify" {
				log.Println("Could not parse Spotify uri.")
				continue
			}

			if strings.Contains(spotifyUri, "local") == true {
				continue
			}

			if reflect.DeepEqual(SpotifyLogin, spotify.Credentials{}) == true {
				log.Println("Could not log in to Spotify.")
				return
			}

			if media.Player != nil && media.Player.StatusCode() != STOPPED {
				switch media.Player.(type) {
				case *SpotifyPlayer:
					media.Player.(*SpotifyPlayer).Outfile = spotifyUri
				default:
					media.Player.Stop(-1)

					media.Player = &SpotifyPlayer{
						Outfile:       spotifyUri,
						KillSwitch:    make(chan int),
						TrackInfo:     make(chan *PlaylistEntry),
						ChangeTrack:   make(chan bool),
						PauseTrack:    make(chan bool),
						ResumeTrack:   make(chan bool),
						StopTrack:     make(chan bool),
						ParsePlaylist: make(chan bool),
						TrackResults:  make(chan []string),
					}
				}
			} else {
				media.Player = &SpotifyPlayer{
					Outfile:       spotifyUri,
					KillSwitch:    make(chan int),
					TrackInfo:     make(chan *PlaylistEntry),
					ChangeTrack:   make(chan bool),
					PauseTrack:    make(chan bool),
					ResumeTrack:   make(chan bool),
					StopTrack:     make(chan bool),
					ParsePlaylist: make(chan bool),
					TrackResults:  make(chan []string),
				}
			}

			if strings.Contains(spotifyUri, "playlist") == true {
				go func() {
					media.Player.(*SpotifyPlayer).ParsePlaylist <- true
					media.Playlist = append(media.Playlist[0:index], <-media.Player.(*SpotifyPlayer).TrackResults...)
				}()
				//log.Println(media.Playlist)
			}

			go func() {
				select {
				case media.Metadata = <-media.Player.(*SpotifyPlayer).TrackInfo:
					log.Println("Media changed.")

					media.MediaChanged <- true
					log.Println("Internal MediaChanged event sent.")
				}
			}()

			go media.Player.Play()

		default:
			outfileChan := make(chan string)

			go YoutubeDl(url, outfileChan)
			outfile := <-outfileChan

			if outfile == "" {
				log.Println("Youtube-dl could not find video link.")
			} else {
				if media.Player.StatusCode() != STOPPED {
					media.Player.Stop(-1)
				}

				media.Player = &OmxPlayer{Outfile: outfile, KillSwitch: make(chan int)}

				go media.Player.Play()
			}
		}

		switch media.Player.ReturnCode() {
		case -1:
			break
		case 1:
			continue
		}

	}

	if media.Player.StatusCode() != STOPPED {
		media.Player.Stop(-1)
	}

}

func (media *Media) Status(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(media.StatusBuilder())
}

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
