package picast

import (
	"log"
	"os/exec"
	"strconv"
	"os"
	"time"
)

func YoutubeDl(entry PlaylistEntry) (string) {
	outfile := "res/cache/" + strconv.Itoa(entry.Id) + ".mp4"
	cmd := exec.Command("youtube-dl", "--no-part", "-o", outfile, entry.Url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	// While outFile does not exist, sleep
	// Return outFileName

	for {
		_, err := os.Stat(outfile);

		if err == nil {
			break
		}

		time.Sleep(1000 * time.Millisecond)

	}

	time.Sleep(3000 * time.Millisecond)
	return outfile
}
