package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	syblog "github.com/syb-devs/gotools/log"
)

var log = syblog.New(os.Stderr)

var errInvalidRepo = errors.New("no valid repository found in the specified path")

func main() {
	defer errHandler()

	var verbose, debug bool
	var path, remote, branch, initFrom string
	var interval time.Duration

	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.BoolVar(&debug, "debug", false, "more verbose output, for debugging purposes")
	flag.StringVar(&path, "path", "", "the path of the repository")
	flag.StringVar(&remote, "remote", "origin", "the name of the remote to sync with")
	flag.StringVar(&branch, "branch", "master", "the name of the remote branch to sync")
	flag.StringVar(&initFrom, "initfrom", "", "the repository remote URL ")
	flag.DurationVar(&interval, "interval", time.Minute, "the time interval for syncing (1m, 35s, 2m3s, 500ms...)")

	flag.Parse()

	if debug {
		log.SetLevel(syblog.LevelDebug)
	} else if verbose {
		log.SetLevel(syblog.LevelInfo)
	} else {
		log.SetLevel(syblog.LevelError)
	}

	log.Info(fmt.Sprintf("repository path: %s", path))
	log.Info(fmt.Sprintf("remote: %s", remote))
	log.Info(fmt.Sprintf("branch: %s", branch))
	log.Info(fmt.Sprintf("sync interval: %s", interval))

	repo := NewRepo(path, remote, branch)
	valid, err := repo.Valid()
	check(err)

	if !valid {
		if initFrom == "" {
			abort(errInvalidRepo)
		}
		exists, err := dirExists(path)
		check(err)

		if !exists {
			log.Info(fmt.Sprintf("creating directory for cloning: %s", path))
			err = os.MkdirAll(path, 0700)
			check(err)
		}
		repo.InitFrom(initFrom)
	}

	tc := time.Tick(interval)
	for {
		select {
		case <-tc:
			err := repo.Sync()
			if err != nil {
				log.Error(fmt.Sprintf("error syncing repo: %v\n", err))
			}
		}
	}
}

func check(err error) {
	if err != nil {
		abort(err)
	}
}

func abort(err interface{}) {
	log.Critical(fmt.Sprintf("aborting. error: %v", err))
	os.Exit(1)
}

func errHandler() {
	if err := recover(); err != nil {
		abort(err)
	}
}

func fileExists(path string) (bool, error) {
	return pathExists(path, false)
}

func dirExists(path string) (bool, error) {
	return pathExists(path, true)
}

func pathExists(path string, dir bool) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if dir {
			return fi.IsDir(), nil
		}
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
