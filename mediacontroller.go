package picast

import (
	"log"
	"strings"
	//"strconv"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	//"os"
	//"io/ioutil"
	//"os/exec"
)

// Plays current entry. After completion, checks for more
// playlist entries and plays them
// Gets currently selected item from sidebar
func (api *Api) PlayAll(w rest.ResponseWriter, r *rest.Request) {
	/*// start from top of playlist

	switch {
	case api.CurrentMedia.Player == nil:
		break
	case api.CurrentMedia.Player.Started() == 1:
		api.CurrentMedia.Player.Stop(-1)
	}

	for api.CurrentMedia.Metadata = api.GetFirst(); *api.CurrentMedia.Metadata != (PlaylistEntry{}); api.CurrentMedia.Metadata = api.GetNext() {
		if strings.Contains(api.CurrentMedia.Metadata.Url, "youtube") {
			api.CurrentMedia.Player = &OmxPlayer{Outfile: YoutubeDl(*api.CurrentMedia.Metadata), KillSwitch: make(chan int)}
			// Made an unbuffered kill channel so the end of this loop will block
			// until either an internal or external kill signal is received

			go api.CurrentMedia.Player.Play()

			// Below breaks out of playlist loop and returns if external kill signal was received
			// Otherwise blocks until internal kill signal receive
			if api.CurrentMedia.Player.ReturnCode() == -1 {
				break
			}
		}
	}

	api.CurrentMedia.Metadata = &PlaylistEntry{}
	api.CurrentMedia.Player = nil
	w.WriteJson(&struct{ Server string }{Server: "Finished playlist."})*/
}

func (api *Api) Next(w rest.ResponseWriter, r *rest.Request) {
	if *api.CurrentMedia.Metadata != (PlaylistEntry{}) {
		nextEntry := api.GetNext()
		api.CurrentMedia.Player.Stop(-1)
		api.CurrentMedia.Metadata = nextEntry

		go api.PlayAll(w, r)
	}
}

func (api *Api) Prev(w rest.ResponseWriter, r *rest.Request) {

}

func (media *Media) Play(w rest.ResponseWriter, r *rest.Request) {
	entry := PlaylistEntry{Id: 0, Url: ""}

	err := r.DecodeJsonPayload(&entry)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	switch {
	case entry.Url == "":
		rest.NotFound(w, r)
		return
	case media.Player != nil:
		media.Stop(w, r)

	}

	media.Metadata = &entry
	outfile, err := YoutubeDl(entry)
	if err != nil {
		log.Println("Youtube-dl could not find video link.")
	} else {
		outfile = strings.Trim(outfile, " \n")
		media.Player = &OmxPlayer{Outfile: outfile, KillSwitch: make(chan int, 1)}
		go media.Player.Play()
	}

	w.WriteJson(media.StatusBuilder())

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
