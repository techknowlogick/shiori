# Shiori

[![Build Status](https://cloud.drone.io/api/badges/techknowlogick/shiori/status.svg)](https://cloud.drone.io/techknowlogick/shiori)
[![Go Report Card](https://goreportcard.com/badge/src.techknowlogick.com/shiori)](https://goreportcard.com/report/src.techknowlogick.com/shiori)
[![GoDoc](https://godoc.org/src.techknowlogick.com/shiori?status.svg)](https://godoc.org/src.techknowlogick.com/shiori)
[![GitHub release](https://img.shields.io/github/release-pre/techknowlogick/shiori.svg)](https://github.com/techknowlogick/shiori/releases/latest)

Shiori is a simple bookmarks manager written in Go language. Intended as a simple clone of [Pocket](https://getpocket.com/). You can use it as command line application or as web application. This application is distributed as a single binary, which means it can be installed and used easily.

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

### Is the original Shiori not being updated anymore?

Sadly the original maintainer of Shiori is no longer able to maintain the project anymore. This project was forked during a prior hiatus and changes with the original project means this project couldn't be merged back. Please see the [answer to this question provided by Radhi](https://github.com/go-shiori/shiori/issues/256) for more details about why the original project is no longer maintained.

## License

Shiori is distributed using [MIT license](https://choosealicense.com/licenses/mit/), which means you can use and modify it however you want. However, if you make an enhancement or a bug fix, if possible, please send a pull request.
