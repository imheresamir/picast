package picast

import (
	"log"
	//"os"
	"os/exec"
	//"strconv"
	//"os"
	"strings"
)

func YoutubeDl(url string, result chan string) {
	video_link_bytes, err := exec.Command("youtube-dl", "-g", url).Output()
	if err != nil {
		log.Println(err)
	}

	video_link := string(video_link_bytes[:])

	switch {
	case video_link == "":
		log.Println("Could not find video link")
		result <- ""
	}

	video_link = strings.Trim(video_link, " \n")

	result <- video_link
}
