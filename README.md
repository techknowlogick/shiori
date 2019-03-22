# Shiori

[![Build Status](https://cloud.drone.io/api/badges/techknowlogick/shiori/status.svg)](https://cloud.drone.io/techknowlogick/shiori)
[![Go Report Card](https://goreportcard.com/badge/src.techknowlogick.com/shiori)](https://goreportcard.com/report/src.techknowlogick.com/shiori)
[![GoDoc](https://godoc.org/src.techknowlogick.com/shiori?status.svg)](https://godoc.org/src.techknowlogick.com/shiori)
[![GitHub release](https://img.shields.io/github/release-pre/techknowlogick/shiori.svg)](https://github.com/techknowlogick/shiori/releases/latest)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ftechknowlogick%2Fshiori.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Ftechknowlogick%2Fshiori?ref=badge_shield)

Shiori is a simple bookmarks manager written in Go language. Intended as a simple clone of [Pocket](https://getpocket.com//). You can use it as command line application or as web application. This application is distributed as a single binary, which means it can be installed and used easily.

## Features

- Simple and clean command line interface.
- Basic bookmarks management i.e. add, edit and delete.
- Search bookmarks by their title, tags, url and page content.
- Import and export bookmarks from and to Netscape Bookmark file.
- Portable, thanks to its single binary format and sqlite3 database
- Simple web interface for those who don't want to use a command line app.
- Where possible, by default `shiori` will download a static copy of the webpage in simple text and HTML format, which later can be used as an offline archive for that page.

## Demo

A demo can be found at: https://shiori-demo.techknowlogick.com/ where the user/password are both `demo`. The database will be wiped every so often, so please don't use this for anything other than quick evaluation.

## FAQs

### Why did you make this fork?

My goals for this fork were to change things for my personal preferences.

### What are the changes made to this repo?

Please see the [list of changes](https://github.com/techknowlogick/shiori/issues/82) made to this project compared to the original.

### Is there any contact between the two repos?

Yes, I've let Radhi know about my changes and we will work on trying to get them back into the original project. Due to some changes I made make it harder to create PRs to merge things back into the original.

### Is the original Shiori not being updated anymore?

There was a brief hiatus, however Radhi plans on updating the original project again. There have already been improvments made to the go-readability project by Radhi. Please see the [answer to this question provided by Radhi](https://github.com/RadhiFadlillah/shiori/issues/119#issuecomment-467273449) for more details.

### Are the two projects compatible?

Switching between the two projects is currently possible by replacing your existing binary, however DB schema changes may eventually be made that require a DB migration.

### Should I use this fork instead of the original?

Please support the [original project](https://github.com/RadhiFadlillah/shiori), however if you prefer to use this fork that is OK too.

## License

Shiori is distributed using [MIT license](https://choosealicense.com/licenses/mit/), which means you can use and modify it however you want. However, if you make an enhancement, if possible, please send a pull request to the [original project](https://github.com/RadhiFadlillah/shiori). If you make a bugfix, please send a pull request to both projects.


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ftechknowlogick%2Fshiori.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Ftechknowlogick%2Fshiori?ref=badge_large)
