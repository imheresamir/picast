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
	//"os"
	//"io/ioutil"
	//"os/exec"
	//"image/jpeg"
	//"os"
	//"time"
)

func (media *Media) Init() {

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
	case media.Player != nil:
		media.Stop(w, r)
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

	go media.PlayAll(object.Url)

	w.WriteJson(media.StatusBuilder())

}

func (media *Media) PlayAll(url string) {
	switch {
	case strings.Contains(url, "spotify"):
		spotifyUri := "spotify"

		re := regexp.MustCompile(`https?:\/\/open\.spotify\.com\/(\w+)\/(\w+)|spotify:(\w+):(\w+)`)
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
			return
		}

		if reflect.DeepEqual(SpotifyLogin, spotify.Credentials{}) == true {
			log.Println("Could not log in to Spotify.")
			return
		}

		media.Player = &SpotifyPlayer{
			Outfile:    spotifyUri,
			KillSwitch: make(chan int, 1),
			TrackInfo:  make(chan *PlaylistEntry),
		}

		go func() {
			select {
			case media.Metadata = <-media.Player.(*SpotifyPlayer).TrackInfo:
				log.Println("Media changed.")

				media.MediaChanged <- true
				log.Println("Internal MediaChanged event sent.")
			}
		}()

		media.Player.Play()

		/*log.Println("TEST sleep...")
		time.Sleep(5 * time.Second)*/

	default:
		outfile, err := YoutubeDl(url)
		if err != nil {
			log.Println("Youtube-dl could not find video link.")
		} else {
			media.Player = &OmxPlayer{Outfile: outfile, KillSwitch: make(chan int, 1)}
			go media.Player.Play()
		}
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
	if media.Player != nil && media.Player.StatusCode() > 0 {
		media.Player.Stop(-1)
		media.Player = nil
	}

	w.WriteJson(media.StatusBuilder())
}
