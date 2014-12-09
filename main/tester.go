package main

import (
	"gopkg.in/qml.v1"
	"os"
	"os/exec"
	"runtime/debug"
)

func main() {
	go func() {
		cmd := exec.Command("youtube-dl", "-g", "https://www.youtube.com/watch?v=B5_2GJweQiw")
		//cmd := exec.Command("echo", "hello world")
		cmd.Stdout = os.Stdout
		cmd.Start()

		debug.PrintStack()

		cmd.Wait()

		debug.PrintStack()

		//cmd.Process.Kill()
		//cmd.Process.Wait()
	}()

	qml.Run(run)
}

func run() error {
	engine := qml.NewEngine()

	component, err := engine.LoadFile("res/main.qml")
	if err != nil {
		return err
	}

	win := component.CreateWindow(nil)
	win.Show()

	debug.PrintStack()

	win.Wait()

	return nil
}
