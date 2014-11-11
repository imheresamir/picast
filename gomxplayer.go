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

func (video *OmxPlayer) ReturnCode() int {
	return <-video.KillSwitch
}

func (video *OmxPlayer) Started() int {
	return video.ThreadStarted
}

func (video *OmxPlayer) Play() {
	if video.Outfile != "" {
		//video.Started = 0
		video.ThreadStarted = 0

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
				//video.Started = 1
				break
			}

		}

		outString := string(outBytes[:])

		video.Duration, err = strconv.ParseInt(reg.FindStringSubmatch(outString)[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		// Get Status
		reg = regexp.MustCompile("Paused: (.+)")

		outBytes, err = exec.Command("util/dbuscontrol.sh", "status").Output()
		if err != nil {
			log.Fatal(err)
		}

		outString = string(outBytes[:])

		paused := reg.FindStringSubmatch(outString)[1]
		if err != nil {
			log.Fatal(err)
		}

		if paused == "false" {
			video.Status = 1
		} else if paused == "true" {
			video.Status = 0
		}

		// Launch goroutine
		video.ThreadStarted = 1
		go video.WatchPosition()
	}
}

func (video *OmxPlayer) TogglePause() {
	if video.ThreadStarted == 1 {
		cmd := exec.Command("util/dbuscontrol.sh", "pause")

		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		if video.Status == 1 {
			video.Status = 0
		} else if video.Status == 0 {
			video.Status = 1
		}
	}
}

// Stop method can be called internally from WatchPosition with kill signal 1 on normal video end
// OR externally from Api or Media methods with kill signal -1
func (video *OmxPlayer) Stop(signal int) {
	if video.ThreadStarted == 1 {
		/*for ; video.ThreadStarted != 1; {
			log.Println("Sleeping in Kill")
			time.Sleep(500 * time.Millisecond)
		}*/
		video.ThreadStarted = 0
		time.Sleep(500 * time.Millisecond)

		cmd := exec.Command("killall", "omxplayer")
		cmd.Run()

		cmd = exec.Command("killall", "/usr/bin/omxplayer.bin")
		cmd.Run()

		video.Outfile = ""
		//video.Started = 0
		video.ThreadStarted = 0
		video.Status = 0
		video.Duration = 0
		video.Position = 0
		//video = &OmxPlayer{}

		video.KillSwitch <- signal

	}
}

func (video *OmxPlayer) WatchPosition() {
	for {
		if video.ThreadStarted != 1 {
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
				log.Println("DBus Error")
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

		//log.Println("Duration: " + strconv.Itoa(video.Duration))
		//log.Println("Position: " + strconv.Itoa(video.Position))

	}
}
