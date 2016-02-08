Hikikomori Project
==================

This project permit you to download Manga chapters and synchronise it
with your MyAnimelist.


## Build Project


```
go build
```

## Usage

```
Usage:
  hikikomori [command]

Available Commands:
  syncMal     Sync MyAnimelist manga list with config file
  download    Download new chapters
  read        Read new downloaded chapter
  help        Help about any command

Flags:
  -h, --help=false: help for hikikomori


Use "hikikomori [command] --help" for more information about a command.
```

## Dependances

* github.com/ghodss/yaml
  To permit transformation beetween yaml and json
* github.com/nstratos/go-myanimelist/mal
  Golang Myanimelist API
* github.com/pierrre/mangadownloader
  Downloader Manga scans
