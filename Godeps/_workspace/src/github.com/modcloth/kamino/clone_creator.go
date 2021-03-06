package kamino

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/modcloth/go-fileutils"
)

type clone struct {
	*Genome
	workdir string
}

func (creator *clone) cachePath() string {
	return fmt.Sprintf("%s/%s/%s", creator.workdir, creator.Account, creator.Repo)
}

func (creator *clone) cloneCacheIfAvailable() (string, error) {
	if err := creator.updateToRef(creator.cachePath()); err != nil {
		return creator.cloneNoCache()
	}

	return creator.cachePath(), nil
}

func (creator *clone) cloneForceCache() (string, error) {
	if err := creator.updateToRef(creator.cachePath()); err != nil {
		return "", err
	}

	return creator.cachePath(), nil
}

func (creator *clone) cloneCreateCache() (string, error) {
	if err := creator.cloneRepo(creator.cachePath()); err != nil {
		return "", err
	}

	return creator.cachePath(), nil
}

func (creator *clone) cloneNoCache() (string, error) {
	uuid, err := nextUUID()
	if err != nil {
		return "", err
	}

	clonePath := fmt.Sprintf("%s/%s/%s", creator.workdir, creator.Account, uuid)

	if err = creator.cloneRepo(clonePath); err != nil {
		return "", err
	}

	return clonePath, nil
}

func (creator *clone) cloneRepo(dest string) error {
	git, err := fileutils.Which("git")
	if err != nil {
		return err
	}

	buff := &bytes.Buffer{}

	repoURL := &url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   fmt.Sprintf("%s/%s", creator.Account, creator.Repo),
	}

	if creator.APIToken != "" {
		repoURL.User = url.User(creator.APIToken)
	}

	var cloneCmd *exec.Cmd

	if creator.Depth == "" {
		cloneCmd = &exec.Cmd{
			Path:   git,
			Args:   []string{"git", "clone", repoURL.String(), dest},
			Stderr: buff,
		}

	} else {
		cloneCmd = &exec.Cmd{
			Path: git,
			Args: []string{
				"git", "clone", "--depth", creator.Depth, repoURL.String(), dest,
			},
			Stderr: buff,
		}
	}

	if err := cloneCmd.Run(); err != nil {
		Logger.WithFields(logrus.Fields{
			"account":            creator.Account,
			"cache_method":       creator.UseCache,
			"depth":              creator.Depth,
			"ref":                creator.Ref,
			"repo":               creator.Repo,
			"api_token_provided": creator.APIToken != "",
			"go_error":           err,
			"stdout":             buff.String(),
		}).Error("error running clone command")

		return err
	}

	buff.Reset()
	checkoutCmd := &exec.Cmd{
		Path:   git,
		Dir:    dest,
		Args:   []string{"git", "checkout", "--force", creator.Ref},
		Stderr: buff,
	}

	if err := checkoutCmd.Run(); err != nil {
		Logger.WithFields(logrus.Fields{
			"account":            creator.Account,
			"cache_method":       creator.UseCache,
			"depth":              creator.Depth,
			"ref":                creator.Ref,
			"repo":               creator.Repo,
			"api_token_provided": creator.APIToken != "",
			"go_error":           err,
			"stdout":             buff.String(),
		}).Error("error running checkout command")

		return err
	}

	return nil
}

func (creator *clone) updateToRef(dest string) error {
	/*
		workflow as follows:
			git clean -d --force --quiet
			git fetch --prune --quiet
			git checkout --force --quiet <ref>
			git symbolic-ref --quiet HEAD && git pull --rebase --quiet
	*/
	git, err := fileutils.Which("git")
	if err != nil {
		return err
	}

	buff := &bytes.Buffer{}

	cmds := []*exec.Cmd{
		&exec.Cmd{
			Args: []string{"git", "clean", "-d", "--force"},
		},
		&exec.Cmd{
			Args: []string{"git", "fetch", "--prune"},
		},
		&exec.Cmd{
			Args: []string{"git", "checkout", "--force", creator.Ref},
		},
	}

	for _, cmd := range cmds {
		cmd.Path = git
		cmd.Dir = dest
		cmd.Stderr = buff

		if err := cmd.Run(); err != nil {
			Logger.WithFields(logrus.Fields{
				"account":            creator.Account,
				"cache_method":       creator.UseCache,
				"depth":              creator.Depth,
				"ref":                creator.Ref,
				"repo":               creator.Repo,
				"api_token_provided": creator.APIToken != "",
				"go_error":           err,
				"stdout":             buff.String(),
			}).Errorf("error running command %q", strings.Join(cmd.Args, " "))

			return err
		}

		buff.Reset()
	}

	detectBranch := &exec.Cmd{
		Path: git,
		Dir:  dest,
		Args: []string{"git", "symbolic-ref", "HEAD"},
	}

	// no error => we are on a proper branch (as opposed to a detached HEAD)
	if err := detectBranch.Run(); err == nil {
		pullRebase := &exec.Cmd{
			Path:   git,
			Dir:    dest,
			Args:   []string{"git", "pull", "--rebase"},
			Stderr: buff,
		}

		if err = pullRebase.Run(); err != nil {
			Logger.WithFields(logrus.Fields{
				"account":            creator.Account,
				"cache_method":       creator.UseCache,
				"depth":              creator.Depth,
				"ref":                creator.Ref,
				"repo":               creator.Repo,
				"api_token_provided": creator.APIToken != "",
				"go_error":           err,
				"stdout":             buff.String(),
			}).Errorf("error running command %q", strings.Join(pullRebase.Args, " "))

			return err
		}
	}

	return nil
}
