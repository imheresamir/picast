package picast

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

/*func main() {
	video := OmxPlayer{}
	video.Play("tmp.mp4")

	for {
		log.Println("Started: " + strconv.Itoa(video.Started))
		log.Println("Outfile: " + video.Outfile)
		log.Println("Duration: " + strconv.Itoa(video.Duration))
		log.Println("Status: " + strconv.Itoa(video.Status))
		log.Println("Position: " + strconv.Itoa(video.Position))
		log.Println()
		time.Sleep(5000 * time.Millisecond)
	}
}*/

func (video *OmxPlayer) StatusCode() int {
	return video.Status
}

func (video *OmxPlayer) ReturnCode() int {
	return <-video.KillSwitch
}

func (video *OmxPlayer) Play() {
	if video.Outfile != "" {
		video.Status = 1

		// Start video
		cmd := exec.Command("omxplayer", "-I", video.Outfile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		// Initialize OmxPlayer struct

		// Get Duration
		reg := regexp.MustCompile("Duration: (.+)")
		var outBytes []byte

		// Keep trying until Omxplayer connects to DBus
		for {
			outBytes, err = exec.Command("util/dbuscontrol.sh", "status").Output()

			if err == nil {
				video.Status = 3

				break
			} else {
				//log.Println("DBus error, stuck in for loop")
			}

		}

		outString := string(outBytes[:])

		video.Duration, err = strconv.ParseInt(reg.FindStringSubmatch(outString)[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		// Launch goroutine
		go video.WatchPosition()
	}
}

func (video *OmxPlayer) TogglePause() {
	if video.Status > 1 {

		for {
			cmd := exec.Command("util/dbuscontrol.sh", "pause")

			err := cmd.Run()
			if err == nil {
				break
			} else {
				//log.Println("Pause error")
			}

		}

		if video.Status == 3 {
			video.Status = 2
		} else if video.Status == 2 {
			video.Status = 3
			go video.WatchPosition()
		}
	}
}

// Stop method can be called internally from WatchPosition with kill signal 1 on normal video end
// OR externally from Api or Media methods with kill signal -1
func (video *OmxPlayer) Stop(signal int) {
	if video.Status > 0 {
		time.Sleep(500 * time.Millisecond)

		cmd := exec.Command("killall", "omxplayer", "omxplayer.bin")
		cmd.Run()

		log.Println("Video stopped.")

		os.Remove(video.Outfile)

		video.Status = 0
		video.KillSwitch <- signal

	}
}

func (video *OmxPlayer) WatchPosition() {
	for {
		if video.Status < 3 {
			break
		}

		// Get Position
		reg := regexp.MustCompile("Position: (.+)")
		var outBytes []byte

		for {
			var err error
			outBytes, err = exec.Command("util/dbuscontrol.sh", "status").Output()

			if err == nil {
				break
			} else {
				//log.Println("DBus Error")
			}

		}

		outString := string(outBytes[:])

		video.Position, _ = strconv.ParseInt(reg.FindStringSubmatch(outString)[1], 10, 64)
		/*if err != nil {
			log.Fatal(err)
		}*/

		if video.Duration-video.Position <= 350000 {
			video.Stop(1) // Send internal kill signal
		}

		//log.Println("Duration: " + strconv.FormatInt(video.Duration, 10))
		//log.Println("Position: " + strconv.FormatInt(video.Position, 10))

	}
}
