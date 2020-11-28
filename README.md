# DRE

Docker Registry Exploration

## Concept

It connects to docker registries to queries all availables repos and tags.
The tool then proceeds to download each files of **the last 2 layers** for each tags.
It also stores the image config in config.json for further analysis.

## Usage

```sh
$ ./dre -h
Usage: dre <command>

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  ls <url>
    List images from a remote registry

  dump <url> <output-directory>
    Dumps files from images in the registry


$ ./dre dump -h
Usage: dre dump <url> <output-directory>

Dumps files from images in the registry

Arguments:
  <url>                 Registry URL
  <output-directory>    Output directory

Flags:
  -h, --help            Show context-sensitive help.

  -m, --max-layers=2    Max amount of layers to get data from, starting from the last one
  -i, --image=STRING    Image name to filter on
  -t, --tag=STRING      Tag name to filter on
```

```sh
./dre ls https://myregistry.company.com -t latest
```

Will list all images in registry `https://myregistry.company.com` with a `latest` tag.

```sh
./dre dump https://myregistry.company.com ./explore -m2 -t latest
```

Will explore `https://myregistry.company.com` and stores every file of the last `2` layers for all images with a `latest` tag in the `./explore` folder.

## Build

```sh
$ go get ./...
$ go build
```

## Install

```
GOMODULE11=on go get github.com/LeakIX/dre
```