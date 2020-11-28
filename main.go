package main

import (
	"github.com/alecthomas/kong"
	"os"
)

var App struct {
	Ls ListCommand `cmd help:"List images from a remote registry"`
	Dump DumpCommand `cmd help:"Dumps files from images in the registry"`
}

func main() {
	ctx := kong.Parse(&App)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

type DockerConfig struct {
	Config struct {
		Env []string `json:"Env"`
	} `json:"config"`
}

func FileExists(filepath string) bool {
	fileinfo, err := os.Stat(filepath)
	if os.IsNotExist(err) || fileinfo == nil {
		return false
	}
	return true
}
