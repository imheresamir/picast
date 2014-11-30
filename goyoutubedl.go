package picast

import (
	"log"
	//"os"
	"os/exec"
	//"strconv"
	"os"
	"strings"
)

func YoutubeDl(entry PlaylistEntry) (string, error) {
	video_link_bytes, err := exec.Command("youtube-dl", "-g", entry.Url).Output()
	if err != nil {
		log.Println(err)
	}

	video_link := string(video_link_bytes[:])

	switch {
	case video_link == "":
		fallthrough
	case strings.Contains(video_link, "\n"):
		log.Println("Could not find video link")
		return "", os.ErrNotExist
	}

	return video_link, nil
}
