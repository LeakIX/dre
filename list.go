package main

import (
	"fmt"
	"github.com/LeakIX/docker-registry-client/registry"
	"io/ioutil"
	"log"
)

type ListCommand struct {
	Url string `arg name:"url" help:"Registry URL" type:"url"`
	Image string `help:"Image name to filter on" short:"i"`
	Tag string `help:"Tag name to filter on" short:"t"`
}

func (cmd *ListCommand) Run() error {
	log.SetOutput(ioutil.Discard)
	hub, err := registry.NewInsecure(cmd.Url, "", "", log.Flags())
	if err != nil {
		return err
	}
	repos, err := hub.Repositories()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if len(cmd.Image) > 0 && repo != cmd.Image {
			continue
		}
		tags, err := hub.Tags(repo)
		if err != nil {
			return err
		}
		for _, tag := range tags {
			if len(cmd.Tag) > 0 && tag != cmd.Tag {
				continue
			}
			fmt.Printf("%s:%s\n", repo, tag)
		}
	}
	return nil
}