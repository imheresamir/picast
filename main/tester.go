package main

import(
  "github.com/imheresamir/picast"
  "log"
  "strings"
  "time"
)

func main() {
  mainMedia := picast.Media{Metadata: &picast.PlaylistEntry{}}
  log.Println(mainMedia)
  entry := picast.PlaylistEntry{Id: 0, Url: "https://www.youtube.com/watch?v=aS-AEKWXjHw"}
  //log.Println(entry)

  switch {
    case strings.Contains(entry.Url, "youtube"):
      mainMedia.Metadata = &entry
      mainMedia.Player = &picast.OmxPlayer{Outfile: picast.YoutubeDl(entry), KillSwitch: make(chan int, 1)}
      //log.Println(mainMedia.Metadata)
      mainMedia.Player.Play()
      //mainMedia.Player.ReturnCode()
      time.Sleep(5000 * time.Millisecond)
      mainMedia.Player.Stop(-1)
  }
}
