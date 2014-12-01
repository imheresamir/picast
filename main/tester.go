package main

import (
	"github.com/imheresamir/picast"
	"github.com/op/go-libspotify/spotify"
	"log"
	"strings"
	"time"
)

func main() {
	mainMedia := picast.Media{Metadata: &picast.PlaylistEntry{}}
	entry := picast.PlaylistEntry{Id: 0, Url: "spotify:track:5vQIaNoKsT4xkG1j2KFBFs"}
	log.Println(entry)

	switch {
	case strings.Contains(entry.Url, "spotify:"):
		mainMedia.Metadata = &entry
		mainMedia.Player = &picast.SpotifyPlayer{
			Outfile:    entry.Url,
			KillSwitch: make(chan int, 1),
			Login: spotify.Credentials{
				Username: "imheresamir",
				Password: "7Hl4h8uh",
			},
		}
		//log.Println(mainMedia.Metadata)
		go mainMedia.Player.Play()
		//mainMedia.Player.ReturnCode()
		time.Sleep(5000 * time.Millisecond)
		mainMedia.Player.TogglePause()

		time.Sleep(5000 * time.Millisecond)
		mainMedia.Player.Stop(-1)
	}
}
