package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"github.com/LeakIX/docker-registry-client/registry"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

type DumpCommand struct {
	Url string `arg name:"url" help:"Registry URL" type:"url"`
	OutputDirectory string `arg name:"output-directory" help:"Output directory" type:"path"`
	MaxLayers int `help:"Max amount of layers to get data from, starting from the last one" short:"m" default:"2"`
	Image string `help:"Image name to filter on" short:"i"`
	Tag string `help:"Tag name to filter on" short:"t"`
}

func (cmd *DumpCommand) Run() error {
	destDir := path.Clean(cmd.OutputDirectory)
	if FileExists(destDir) {
		panic("Destination incorrect/already exists")
	}
	err := os.MkdirAll(destDir, 0700)
	if err != nil {
		return err
	}
	log.Println("Created " + destDir)
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
			log.Printf("Found %s:%s :\n", repo, tag)
			imageStoreDir := path.Join(destDir, repo, tag)
			err = os.MkdirAll(imageStoreDir, 0700)
			if err != nil {
				return err
			}
			log.Println("Created " + imageStoreDir)
			imageRootStoreDir := path.Join(imageStoreDir, "root")
			err = os.MkdirAll(imageRootStoreDir, 0700)
			if err != nil {
				return err
			}
			log.Println("Created " + imageRootStoreDir)
			manifest, err := hub.ManifestV2(repo, tag)
			if err != nil {
				return err
			}
			var files []string
			for idx, layer := range manifest.Layers {
				if idx < len(manifest.Layers)-cmd.MaxLayers {
					continue
				}
				digestReader, err := hub.DownloadBlob(repo, layer.Digest)
				if err != nil {
					return err
				}

				gzr, err := gzip.NewReader(digestReader)
				if err != nil {
					return err
				}
				tr := tar.NewReader(gzr)
				for {
					header, err := tr.Next()
					if err != nil {
						break
					}
					if header == nil {
						continue
					}
					files = append(files, header.Name)
					target := filepath.Join(imageRootStoreDir, header.Name)
					switch header.Typeflag {

					// if its a dir and it doesn't exist create it
					case tar.TypeDir:
						if _, err := os.Stat(target); err != nil {
							if err := os.MkdirAll(target, 0755); err != nil {
								return err
							}
						}
					// if it's a file create it
					case tar.TypeReg:
						log.Printf("Downloading %s to %s...", header.Name, target)
						// Fallback logic in case the directory was created in an upper layer
						baseDir := path.Dir(target)
						if !FileExists(baseDir) {
							if err := os.MkdirAll(baseDir, 0755); err != nil {
								return err
							}
						}
						f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
						if err != nil {
							return err
						}

						// copy over contents
						if _, err := io.Copy(f, tr); err != nil {
							return err
						}
						// manually close here after each file operation; defering would cause each file close
						// to wait until all operations have completed.
						f.Close()
					}
				}
				digestReader.Close()
				gzr.Close()
			}
			digestReader, err := hub.DownloadBlob(repo, manifest.Config.Digest)
			if err != nil {
				return err
			}
			dockerConfig := &DockerConfig{}
			dockerConfigFile, err := os.Create(path.Join(imageStoreDir, "config.json"))
			if err != nil {
				log.Fatalf("failed creating file: %s", err)
			}
			teeIO := io.TeeReader(digestReader, dockerConfigFile)
			jsonDecoder := json.NewDecoder(teeIO)
			err = jsonDecoder.Decode(dockerConfig)
			if err != nil {
				return err
			}
			digestReader.Close()
			dockerConfigFile.Close()
			if len(dockerConfig.Config.Env) > 0 {
				log.Print("Environment: ")
				log.Println(dockerConfig.Config.Env)
			}
			if len(files) > 0 {
				log.Printf("Stored %d files from last %d layers\n", len(files), cmd.MaxLayers)
			}
		}

	}
	return nil
}


