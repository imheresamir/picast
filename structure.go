package picast

import "database/sql"

type Api struct {
	DB           *sql.DB
	CurrentMedia *Media
}

// Matches database schema
type PlaylistEntry struct {
	Id   int
	Url  string
	Data string
}

type MediaPlayer interface {
	Play()
	TogglePause()
	Stop(int) // pass -1 if calling from external (non-MediaPlayer) method
	ReturnCode() int
	StatusCode() int // 0 = stopped, 1 = loading, 2 = paused, 3 = playing
}

type Media struct {
	Metadata *PlaylistEntry
	Player   MediaPlayer
}

type ServerStatus struct {
	Server string
}

type OmxPlayer struct {
	Outfile string

	Status int

	Duration int64
	Position int64

	KillSwitch chan int // Signal to break out of WatchPosition and clear struct
	// internal stop signal = 1, external stop signal = -1
}
