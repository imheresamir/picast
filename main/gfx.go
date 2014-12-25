package main

import (
	"github.com/imheresamir/picast"
	"gopkg.in/qml.v1"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Display struct {
	CurrentTrack picast.PlaylistEntry
	Update       chan bool
}

func (d *Display) SetTrack(track picast.PlaylistEntry, reply *int) error {
	d.CurrentTrack = track
	d.Update <- true

	return nil
}

func main() {
	qml.Run(run)
}

func run() error {
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
		log.Println("Starting rpc server")
		display := new(Display)
		display.Update = make(chan bool)

		rpc.Register(display)
		rpc.HandleHTTP()

		l, err := net.Listen("tcp", ":8084")
		if err != nil {
			log.Fatal("listen error:", err)
		}

		go http.Serve(l, nil)

		for {
			select {
			case <-display.Update:
				currentTrack.ArtPath = display.CurrentTrack.ArtPath
				currentTrack.Title = display.CurrentTrack.Title
				currentTrack.Artist = display.CurrentTrack.Artist
				currentTrack.Album = display.CurrentTrack.Album

				qml.Changed(currentTrack, &currentTrack.ArtPath)
				qml.Changed(currentTrack, &currentTrack.Title)
				qml.Changed(currentTrack, &currentTrack.Artist)
				qml.Changed(currentTrack, &currentTrack.Album)

				log.Println(currentTrack)
				log.Println("Artfile changed.")
			}
		}
	}()

	win.Wait()

	return nil
}
