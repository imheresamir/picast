// Copyright 2014 Samir Bhatt
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright 2013 Örjan Persson
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package picast

import (
	"fmt"
	"io/ioutil"
	"log"
	//"strings"
	"sync"
	"time"

	"code.google.com/p/portaudio-go/portaudio"
	"github.com/op/go-libspotify/spotify"

	"image/jpeg"
	"os"
)

// PiCast Spotify MediaPlayer

func (spotty *SpotifyPlayer) StatusCode() int {
	return spotty.Status
}

func (spotty *SpotifyPlayer) ReturnCode() int {
	return <-spotty.KillSwitch
}

func (spotty *SpotifyPlayer) Play() {
	if spotty.Status == STOPPED {
		go spotty.SpotifyThread()
	} else {
		spotty.ChangeTrack <- true
	}
}

func (spotty *SpotifyPlayer) TogglePause() {
	/*if spotty.Status < PAUSED {
		return
	}*/

	if spotty.Status == PLAYING {
		spotty.PauseTrack <- true
		spotty.Status = PAUSED
	} else if spotty.Status == PAUSED {
		spotty.ResumeTrack <- true
		spotty.Status = PLAYING
		//go spotty.watchPosition()
	}

	// Prevent funky race conditions
	//time.Sleep(1 * time.Second)
}

func (spotty *SpotifyPlayer) Stop(signal int) {
	spotty.StopTrack <- true

	spotty.Status = STOPPED

	log.Println("Track stopped.")

	spotty.KillSwitch <- signal
}

func (spotty *SpotifyPlayer) SpotifyThread() {

	if spotty.Outfile == "" {
		return
	}

	spotty.Status = LOADING

	appKey, err := ioutil.ReadFile("res/spotify_appkey.key")
	if err != nil {
		log.Println(err)
		return
	}

	log.Print("Opening AudioWriter....")
	Audio, err := newAudioWriter()
	if err != nil {
		log.Println(err)
		return
	}
	defer Audio.Close()
	//log.Println(" done.")

	log.Print("Creating Spotify session....")

	Session, err := spotify.NewSession(&spotify.Config{
		ApplicationKey:   appKey,
		ApplicationName:  "picast",
		CacheLocation:    "res/cache",
		SettingsLocation: "res/cache",
		AudioConsumer:    Audio,

		// Disable playlists to make playback faster
		DisablePlaylistMetadataCache: true,
		InitiallyUnloadPlaylists:     true,
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer Session.Close()
	//log.Println(" done.")

	// Log messages

	debug := true
	if debug {
		go func() {
			for msg := range Session.LogMessages() {
				log.Print(msg)
			}
		}()
	}

	log.Print("Logging in....")

	if err = Session.Login(SpotifyLogin, true); err != nil {
		log.Println(err)
		return
	}

	// Wait for login and expect it to go fine
	err = <-Session.LoggedInUpdates()
	if err != nil {
		log.Println(err)
		return
	}
	defer Session.Logout()

	newTrack := true
	var Player *spotify.Player

	for {
		if newTrack == true {
			// Parse the track
			link, err := Session.ParseLink(spotty.Outfile)
			if err != nil {
				log.Println(err)
				return
			}
			track, err := link.Track()
			if err != nil {
				log.Println(err)
				return
			}

			// Load the track and play it
			track.Wait()
			Player = Session.Player()
			if err := Player.Load(track); err != nil {
				fmt.Println("%#v", err)
			}

			Player.Play()

			if spotty.Status == PAUSED {
				Audio.Resume()
			}

			track.Wait()

			artfile := "cache/art.jpg"
			spotty.wg.Add(1)
			go func() {
				os.Remove("res/" + artfile)
				image, err := track.Album().Cover(spotify.ImageSizeLarge)
				if err != nil {
					fmt.Println(err)
				} else {
					// TODO: Add error checking and refactor
					image.Wait()
					toimg, _ := os.Create("res/" + artfile)
					img, _, _ := image.Decode()
					jpeg.Encode(toimg, img, &jpeg.Options{jpeg.DefaultQuality})
					toimg.Close()

				}

				spotty.Duration = track.Duration()

				//go spotty.watchPosition()

				var artists []string
				for i := 0; i < track.Artists(); i++ {
					artists = append(artists, track.Artist(i).Name())
				}

				spotty.TrackInfo <- &PlaylistEntry{
					Title:   track.Name(),
					Artists: artists,
					Album:   track.Album().Name(),
					ArtPath: artfile,
				}
				log.Println("Spotify sent track info.")

				spotty.wg.Done()
			}()

			newTrack = false
			spotty.Status = PLAYING
		}

		select {
		case <-Session.EndOfTrackUpdates():
			spotty.Stop(1)
		case <-spotty.ChangeTrack:
			log.Println("Changing track")
			if spotty.Status == PLAYING {
				Audio.Pause()
				time.Sleep(50 * time.Millisecond)
				Player.Pause()
				spotty.Status = PAUSED
			}
			Player.Unload()
			newTrack = true
			continue
		case <-spotty.StopTrack:
			spotty.wg.Wait()
			Player.Unload()
			return
		case <-spotty.PauseTrack:
			Audio.Pause()
			time.Sleep(50 * time.Millisecond)
			Player.Pause()
			//time.Sleep(500 * time.Millisecond)
			//Audio.Pause()
			continue
		case <-spotty.ResumeTrack:
			log.Println("Resuming")
			Audio.Resume()
			Player.Play()
			continue
		}
	}
}

func (spotty *SpotifyPlayer) watchPosition() {
	now := time.Now()
	start := now
	initialPos := spotty.Position

	for {
		if spotty.Position >= spotty.Duration || spotty.Status < PLAYING {
			break
		}

		now = <-time.Tick(time.Second / 10)
		spotty.Position = initialPos + now.Sub(start)

		//log.Println("Position: ", spotty.Position, "Duration: ", spotty.Duration)
	}

}

// Core helpers

var (
	// audioInputBufferSize is the number of delivered data from libspotify before
	// we start rejecting it to deliver any more.
	audioInputBufferSize = 8

	// audioOutputBufferSize is the maximum number of bytes to buffer before
	// passing it to PortAudio.
	audioOutputBufferSize = 8192
)

// audio wraps the delivered Spotify data into a single struct.
type audio struct {
	format spotify.AudioFormat
	frames []byte
}

// audioWriter takes audio from libspotify and outputs it through PortAudio.
type audioWriter struct {
	input  chan audio
	quit   chan bool
	pause  chan bool
	resume chan bool
	wg     sync.WaitGroup
}

// newAudioWriter creates a new audioWriter handler.
func newAudioWriter() (*audioWriter, error) {
	w := &audioWriter{
		input:  make(chan audio, audioInputBufferSize),
		quit:   make(chan bool, 1),
		pause:  make(chan bool, 1),
		resume: make(chan bool, 1),
	}

	log.Print("Opening portaudio stream...")
	stream, err := newPortAudioStream()
	if err != nil {
		return w, err
	}
	log.Println(" done.")
	w.wg.Add(1)
	go w.streamWriter(stream)
	return w, nil
}

// Close stops and closes the audio stream and terminates PortAudio.
func (w *audioWriter) Close() error {
	select {
	case w.quit <- true:
	default:
	}
	w.wg.Wait()

	close(w.input)
	close(w.quit)
	close(w.pause)
	close(w.resume)

	return nil
}

func (w *audioWriter) Pause() error {
	w.pause <- true
	return nil
}

func (w *audioWriter) Resume() error {
	w.resume <- true
	return nil
}

// WriteAudio implements the spotify.AudioWriter interface.
func (w *audioWriter) WriteAudio(format spotify.AudioFormat, frames []byte) int {
	select {
	case w.input <- audio{format, frames}:
		return len(frames)
	default:
		return 0
	}
}

// streamWriter reads data from the input buffer and writes it to the output
// portaudio buffer.
func (w *audioWriter) streamWriter(stream *portAudioStream) {
	defer w.wg.Done()
	defer stream.Close()

	buffer := make([]int16, audioOutputBufferSize)
	output := buffer[:]

	for {
		// Wait for input data or signal to quit.
		var input audio
		processAudio := true
		select {
		case input = <-w.input:
		case <-w.quit:
			return
		case <-w.pause:
			if stream != nil {
				//if err := stream.stream.Stop(); err != nil {
				if err := stream.stream.Abort(); err != nil {
					log.Println(err)
				}

				processAudio = false
			}
		case <-w.resume:
			if stream != nil {
				if err := stream.stream.Start(); err != nil {
					log.Println(err)
				}

				processAudio = true
				input = <-w.input
			}

		}

		if processAudio == true {
			// Initialize the audio stream based on the specification of the input format.
			err := stream.Stream(&output, input.format.Channels, input.format.SampleRate)
			if err != nil {
				panic(err)
			}

			// Decode the incoming data which is expected to be 2 channels and
			// delivered as int16 in []byte, hence we need to convert it.
			i := 0
			for i < len(input.frames) {
				j := 0
				for j < len(buffer) && i < len(input.frames) {
					buffer[j] = int16(input.frames[i]) | int16(input.frames[i+1])<<8
					j += 1
					i += 2
				}

				output = buffer[:j]
				stream.Write()
			}
		}
	}
}

// portAudioStream manages the output stream through PortAudio when requirement
// for number of channels or sample rate changes.
type portAudioStream struct {
	device *portaudio.DeviceInfo
	stream *portaudio.Stream

	channels   int
	sampleRate int
}

// newPortAudioStream creates a new portAudioStream using the default output
// device found on the system. It will also take care of automatically
// initialise the PortAudio API.
func newPortAudioStream() (*portAudioStream, error) {
	log.Print("Initializing PortAudio...")
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}
	log.Print(" done.")

	log.Print("Creating PortAudio stream...")
	out, err := portaudio.DefaultHostApi()
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}
	log.Print(" done.")
	return &portAudioStream{device: out.DefaultOutputDevice}, nil
}

// Close closes any open audio stream and terminates the PortAudio API.
func (s *portAudioStream) Close() error {
	if err := s.reset(); err != nil {
		portaudio.Terminate()
		return err
	}
	if err := portaudio.Terminate(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *portAudioStream) reset() error {
	if s.stream != nil {
		if err := s.stream.Stop(); err != nil {
			return err
		}
		if err := s.stream.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Stream prepares the stream to go through the specified buffer, channels and
// sample rate, re-using any previously defined stream or setting up a new one.
func (s *portAudioStream) Stream(buffer *[]int16, channels int, sampleRate int) error {
	if s.stream == nil || s.channels != channels || s.sampleRate != sampleRate {
		if err := s.reset(); err != nil {
			return err
		}

		params := portaudio.HighLatencyParameters(nil, s.device)
		params.Output.Channels = channels
		params.SampleRate = float64(sampleRate)
		params.FramesPerBuffer = len(*buffer)

		stream, err := portaudio.OpenStream(params, buffer)
		if err != nil {
			return err
		}
		if err := stream.Start(); err != nil {
			stream.Close()
			return err
		}

		s.stream = stream
		s.channels = channels
		s.sampleRate = sampleRate
	}
	return nil
}

// Write pushes the data in the buffer through to PortAudio.
func (s *portAudioStream) Write() error {
	return s.stream.Write()
}
