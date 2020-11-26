package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/kycheng/docker-registry-client/registry"
	"github.com/micro/go-micro/util/file"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

var maxLayers = 2

func main() {
	//log.SetOutput(ioutil.Discard)
	// creating target folder, denying if existing
	if len(os.Args) < 3 {
		log.Fatal("2 arguments required, repo and target directory")
	}
	destDir := path.Clean(os.Args[2])
	destDirExists, err := file.Exists(destDir)
	if err != nil || destDirExists {
		log.Fatal("Destination incorrect/already exists", err)
	}
	err = os.MkdirAll(destDir, 0700)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created " + destDir)
	hub, err := registry.NewInsecure(os.Args[1], "", "", log.Flags())
	if err != nil {
		panic(err)
	}
	repos, err := hub.Repositories()
	if err != nil {
		panic(err)
	}
	for _, repo := range repos {
		tags, err := hub.Tags(repo)
		if err != nil {
			panic(err)
		}
		for _, tag := range tags {
			fmt.Printf("Found %s:%s :\n", repo, tag)
			imageStoreDir := path.Join(destDir, repo, tag)
			err = os.MkdirAll(imageStoreDir, 0700)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Created " + imageStoreDir)
			imageRootStoreDir := path.Join(imageStoreDir, "root")
			err = os.MkdirAll(imageRootStoreDir, 0700)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Created " + imageRootStoreDir)
			manifest, err := hub.ManifestV2(repo, tag)
			if err != nil {
				panic(err)
			}
			var files []string
			for idx, layer := range manifest.Layers {
				if idx < len(manifest.Layers)-maxLayers {
					continue
				}
				digestReader, err := hub.DownloadBlob(repo, layer.Digest)
				if err != nil {
					panic(err)
				}

				gzr, err := gzip.NewReader(digestReader)
				if err != nil {
					panic(err)
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
								panic(err)
							}
						}
					// if it's a file create it
					case tar.TypeReg:
						log.Printf("Downloading %s...", header.Name)
						f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
						if err != nil {
							panic(err)
						}

						// copy over contents
						if _, err := io.Copy(f, tr); err != nil {
							panic(err)
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
				panic(err)
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
				panic(err)
			}
			digestReader.Close()
			dockerConfigFile.Close()
			if len(dockerConfig.Config.Env) > 0 {
				log.Print("Environment: ")
				log.Println(dockerConfig.Config.Env)
			}
			if len(files) > 0 {
				log.Printf("Stored %d files from last %d layers\n", len(files), maxLayers)
			}
			fmt.Println()
		}

	}
}

type DockerConfig struct {
	Config struct {
		Env []string `json:"Env"`
	} `json:"config"`
}
