package picast

import "database/sql"

type Api struct {
  DB *sql.DB
  CurrentMedia *Media
}

// Matches database schema
type PlaylistEntry struct {
  Id			int
  Url			string
  //Played	int
}

type MediaPlayer interface {
  Play()
  TogglePause()
  Stop(int) // pass -1 if calling from external (non-MediaPlayer) method
  Started() int // 0 = stopped, 1 = started
  ReturnCode() int
}

type Media struct {
  Metadata *PlaylistEntry
  Player MediaPlayer
}

type OmxPlayer struct {
  Outfile string

  ThreadStarted int // 0 = stopped, 1 = started: OmxPlayer is initialized and WatchPosition is active
  // ThreadStarted is now being used to stop WatchPosition from Stop()
  Status int // 0 = paused, 1 = playing

  Duration int64
  Position int64

  KillSwitch chan int // Signal to break out of WatchPosition and clear struct
  // internal stop signal = 1, external stop signal = -1
}
