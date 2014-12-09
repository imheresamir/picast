package picast

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	OmxProcess *exec.Cmd
)

func (video *OmxPlayer) StatusCode() int {
	return video.Status
}

func (video *OmxPlayer) ReturnCode() int {
	return <-video.KillSwitch
}

func (video *OmxPlayer) Play() {

	if video.Outfile == "" {
		return
	}

	video.Status = LOADING

	// Start video
	OmxProcess = exec.Command("omxplayer", "-I", video.Outfile)
	//OmxProcess = exec.Command("vlc", video.Outfile)
	OmxProcess.Stdout = os.Stdout
	OmxProcess.Stderr = os.Stderr

	err := OmxProcess.Start()
	if err != nil {
		log.Println(err)
		video.Status = STOPPED
		return
	}

	// Initialize OmxPlayer struct

	// Get Duration
	reg := regexp.MustCompile("Duration: (.+)")
	var outBytes []byte

	// Keep trying until Omxplayer connects to DBus
	for {
		outBytes, err = exec.Command("util/dbuscontrol.sh", "status").Output()

		if err == nil {
			video.Status = PLAYING

			break
		} else {
			//log.Println("DBus error, stuck in for loop")
		}

	}

	outString := string(outBytes[:])

	d, err := strconv.ParseInt(reg.FindStringSubmatch(outString)[1], 10, 64)
	if err != nil {
		log.Println(err)
	} else {
		video.Duration = time.Duration(d) * time.Microsecond
	}

	// Launch goroutine
	go video.watchPosition()
}

func (video *OmxPlayer) TogglePause() {
	if video.Status < PAUSED {
		return
	}

	for {
		cmd := exec.Command("util/dbuscontrol.sh", "pause")

		err := cmd.Run()
		if err == nil {
			break
		} else {
			//log.Println("Pause error")
		}

	}

	if video.Status == PLAYING {
		video.Status = PAUSED
	} else if video.Status == PAUSED {
		video.Status = PLAYING
		go video.watchPosition()
	}
}

// Stop method can be called internally from WatchPosition with kill signal 1 on normal video end
// OR externally from Api or Media methods with kill signal -1
func (video *OmxPlayer) Stop(signal int) {
	if video.Status == STOPPED {
		return
	}

	for {
		cmd := exec.Command("util/dbuscontrol.sh", "stop")

		err := cmd.Run()
		if err == nil {
			break
		}

	}

	if OmxProcess != nil {
		OmxProcess.Process.Kill()
		OmxProcess.Process.Wait()
		OmxProcess = nil
	}

	log.Println("Video stopped.")

	video.Status = STOPPED
	video.KillSwitch <- signal

}

func (video *OmxPlayer) watchPosition() {
	for {
		if video.Status < PLAYING {
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

		p, err := strconv.ParseInt(reg.FindStringSubmatch(outString)[1], 10, 64)
		if err != nil {
			log.Println(err)
		} else {
			video.Position = time.Duration(p) * time.Microsecond
		}

		if video.Duration-video.Position <= 350000 {
			video.Stop(1) // Send internal kill signal
		}

		//log.Println("Duration: " + strconv.FormatInt(video.Duration, 10))
		//log.Println("Position: " + strconv.FormatInt(video.Position, 10))

	}
}
