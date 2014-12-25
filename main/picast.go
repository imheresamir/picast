package main

import (
	"flag"
	"github.com/imheresamir/picast"
	"github.com/op/go-libspotify/spotify"
	"log"
	"net/rpc"
	"os/exec"
)

func main() {
	picast.SpotifyLogin = spotify.Credentials{}
	flag.StringVar(&picast.SpotifyLogin.Username, "username", "", "Username")
	flag.StringVar(&picast.SpotifyLogin.Password, "password", "", "Password")
	flag.Parse()

	trackUpdate := make(chan picast.PlaylistEntry)
	go picast.RunServer(trackUpdate)

	cmd := exec.Command("./gfx")
	err := cmd.Start()
	if err != nil {
		log.Println("Could not start graphics thread:", err)
	}
	defer cmd.Process.Kill()

	for {
		select {
		case newTrack := <-trackUpdate:
			client, err := rpc.DialHTTP("tcp", "localhost:8084")
			if err != nil {
				log.Fatal("dialing:", err)
			}

			// Synchronous call

			err = client.Call("Display.SetTrack", newTrack, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
