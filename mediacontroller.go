package picast

import (
	//"log"
	"strings"
	//"strconv"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	"os"
	"os/exec"
)

// Plays current entry. After completion, checks for more
// playlist entries and plays them
// Gets currently selected item from sidebar
func (api *Api) PlayAll(w rest.ResponseWriter, r *rest.Request) {
	// start from top of playlist

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
			// Otherwise continues after internal kill signal receive
			if api.CurrentMedia.Player.ReturnCode() == -1 {
				return
			}
		}
	}

	api.CurrentMedia.Metadata = &PlaylistEntry{}
	w.WriteJson(&struct{ Server string }{Server: "Finished playlist."})
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
		return
	}

	switch {
	case entry.Url == "":
		rest.NotFound(w, r)
		return
	case media.Player != nil:
		//log.Println("In case 2")
		media.Stop(w, r)

		fallthrough
	case strings.Contains(entry.Url, "youtube"):
		//log.Println("In case 3")
		media.Metadata = &entry
		media.Player = &OmxPlayer{Outfile: YoutubeDl(entry), KillSwitch: make(chan int, 1)}
		// Made a buffered kill channel so the internal kill signal won't block

		go media.Player.Play()
		w.WriteJson(&struct{ Server string }{Server: "Unsaved Youtube media playing."})
	}
}

func (media *Media) TogglePause(w rest.ResponseWriter, r *rest.Request) {
	switch {
	case media.Player == nil:
		return
	}

	if media.Player.Started() == 1 {
		media.Player.TogglePause()
	}

	w.WriteJson(&struct{ Server string }{Server: "Media (un)paused."})
}

func (media *Media) Stop(w rest.ResponseWriter, r *rest.Request) {
	switch {
	case media.Player == nil:
		return
	case true: //case media.Player.Started() == 1:
		media.Player.Stop(-1)

		fallthrough
	case strings.Contains(media.Metadata.Url, "youtube"):
		cmd := exec.Command("killall", "youtube-dl")
		cmd.Run()

		media.Player = nil
	}

	w.WriteJson(&struct{ Server string }{Server: "Media stopped."})
}
