package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
)

type repo struct {
	Path   string
	Remote string
	Branch string
}

// NewRepo returns a new repo struct
func NewRepo(path, remote, branch string) *repo {
	return &repo{
		Path:   path,
		Remote: remote,
		Branch: branch,
	}
}

func (r *repo) Valid() (bool, error) {
	return dirExists(path.Join(r.Path, ".git"))
}

// InitFrom clones the remote repository in the local repository's path
func (r *repo) InitFrom(url string) error {
	log.Info(fmt.Sprintf("cloning %s into %s", url, r.Path))
	_, err := r.run("clone", url, ".")
	if err != nil {
		return err
	}
	log.Info("clone finished")
	return nil
}

// Fetch fetches incoming changes from the remote branch
func (r *repo) Fetch(remote, branch string) error {
	log.Debug(fmt.Sprintf("fetching %s %s", remote, branch))
	_, err := r.run("fetch", remote, branch)
	return err
}

// Checks out the given revision
func (r *repo) Checkout(revision string) error {
	log.Info(fmt.Sprintf("checking out %s", revision))
	_, err := r.run("checkout", revision)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("synced to revision %s", revision))
	return nil
}

func (r *repo) Revision() (string, error) {
	return r.revision("HEAD")
}

// RemoteRevision returns the latest revision of the remote repository.
func (r *repo) RemoteRevision() (string, error) {
	return r.revision(r.Remote + "/" + r.Branch)
}

func (r *repo) revision(target string) (string, error) {
	log.Debug(fmt.Sprintf("getting revision for %s", target))
	bs, err := r.run("rev-parse", target)
	if err != nil {
		return "", err
	}
	rev := string(bytes.Trim(bs, "\n"))
	log.Debug(fmt.Sprintf("revision for %s is %s", target, rev))
	return rev, nil
}

// Sync checks if the remote revision is different than the local one, updating the repository if it is.
func (r *repo) Sync() error {
	err := r.Fetch(r.Remote, r.Branch)
	if err != nil {
		return err
	}
	locRev, err := r.Revision()
	if err != nil {
		return err
	}
	remRev, err := r.RemoteRevision()
	if err != nil {
		return err
	}
	if locRev != remRev {
		log.Info("the repo is out of sync with remote")
		return r.Checkout(remRev)
	}
	log.Debug("the repo is in sync with remote")
	return nil
}

func (r *repo) run(arg ...string) ([]byte, error) {
	cmdArgs := make([]string, 0, len(arg)+2)
	cmdArgs = append(cmdArgs, "-C", r.Path)
	cmdArgs = append(cmdArgs, arg...)

	cmdStr := fmt.Sprintf("git %v", cmdArgs)
	log.Debug(fmt.Sprintf("running command: %s", cmdStr))

	bs, err := exec.Command("git", cmdArgs...).CombinedOutput()
	if err != nil {
		return bs, &cmdErr{err, cmdStr, bs}
	}
	return bs, err
}

type cmdErr struct {
	err    error
	cmd    string
	cmdOut []byte
}

func (e *cmdErr) Error() string {
	return fmt.Sprintf("%v \nwhile running command [%s] \noutput was [%s]", e.err, e.cmd, e.cmdOut)
}
