package picast

import (
	"github.com/op/go-libspotify/spotify"
	"time"
)

var (
	MainMedia Media
)

// Matches database schema
type PlaylistEntry struct {
	Title   string
	Artist  string
	Album   string
	ArtPath string
	Url     string
}

type ServerObject struct {
	Url  string
	Data interface{}
}

type Media struct {
	Metadata     *PlaylistEntry
	Player       MediaPlayer // Eventually change to anonymous interface
	MediaChanged chan bool
	MediaAdded   chan bool
	Playlist     []PlaylistEntry
	CurrentIndex int
}

type ServerStatus struct {
	Server string
	Url    string
}

type MediaPlayer interface {
	Play()
	TogglePause()
	Stop(int) // pass -1 if calling from external (non-MediaPlayer) method
	ReturnCode() int
	StatusCode() int // 0 = stopped, 1 = loading, 2 = paused, 3 = playing
}

const (
	STOPPED = 0
	LOADING = 1
	PAUSED  = 2
	PLAYING = 3
)

type OmxPlayer struct {
	Outfile string

	Status int

	Duration time.Duration
	Position time.Duration

	KillSwitch chan int // Signal to break out of WatchPosition and clear struct
	// internal stop signal = 1, external stop signal = -1
}

type SpotifyPlayer struct {
	Outfile string

	Status int

	Duration time.Duration
	Position time.Duration

	KillSwitch chan int // Signal to break out of WatchPosition and clear struct
	// internal stop signal = 1, external stop signal = -1

	//TrackInfo chan *PlaylistEntry

	ChangeTrack chan bool
	StopTrack   chan bool
	PauseTrack  chan bool
	ResumeTrack chan bool
	//ParsePlaylist chan bool
	//TrackResults  chan []string
}

var (
	SpotifyLogin spotify.Credentials
)
