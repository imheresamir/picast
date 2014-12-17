package main

import (
	//"errors"
	"flag"
	"github.com/imheresamir/picast"
	"github.com/op/go-libspotify/spotify"
	"gopkg.in/qml.v1"
	"log"
	"math/rand"
	"strconv"
	//"time"
)

func main() {
	picast.SpotifyLogin = spotify.Credentials{}
	flag.StringVar(&picast.SpotifyLogin.Username, "username", "", "Username")
	flag.StringVar(&picast.SpotifyLogin.Password, "password", "", "Password")
	flag.Parse()

	MainDisplay := Display{}

	MainDisplay.Init()

}

type Display struct {
	Update chan *picast.PlaylistEntry
}

func (d *Display) Init() error {
	d.Update = make(chan *picast.PlaylistEntry)
	err := qml.Run(d.run)

	return err
}

func (d *Display) run() error {
	engine := qml.NewEngine()

	currentTrack := &picast.PlaylistEntry{}
	engine.Context().SetVar("currentTrack", currentTrack)

	component, err := engine.LoadFile("res/main.qml")
	if err != nil {
		return err
	}
	win := component.CreateWindow(nil)
	win.Show()

	go func() {
		rand.Seed(42)
		for {
			select {
			case newTrack := <-d.Update:
				currentTrack.ArtPath = "" + newTrack.ArtPath + "?id=" + strconv.Itoa(rand.Intn(10000))
				currentTrack.Title = newTrack.Title
				currentTrack.Artist = newTrack.Artist
				currentTrack.Album = newTrack.Album

				qml.Changed(currentTrack, &currentTrack.ArtPath)
				qml.Changed(currentTrack, &currentTrack.Title)
				qml.Changed(currentTrack, &currentTrack.Artist)
				qml.Changed(currentTrack, &currentTrack.Album)

				//log.Println(currentTrack.Title + " " + currentTrack.Album)
				log.Println("Artfile changed.")
			}
		}
	}()

	go picast.RunServer(d.Update)

	win.Wait()

	return nil
}
