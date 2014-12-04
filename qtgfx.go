package picast

import (
	//"errors"
	"gopkg.in/qml.v1"
	"log"
	"math/rand"
	"strconv"
	//"time"
)

/*func main() {
	Update = make(chan bool)

	go func() {
		time.Sleep(10 * time.Second)
		Update <- true
	}()

	Init()

}*/

func (d *Display) Init() error {
	d.Update = make(chan *PlaylistEntry)
	err := qml.Run(d.run)

	return err
}

func (d *Display) run() error {
	engine := qml.NewEngine()

	currentTrack := &PlaylistEntry{}
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
				qml.Changed(currentTrack, &currentTrack.ArtPath)
				log.Println(currentTrack.ArtPath)
				log.Println("Artfile changed.")
			}
		}
	}()

	win.Wait()

	return nil
}
