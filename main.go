package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Repo type hold common metadata about the remote repository
type Repo struct {
	// url like https://github.com/maikdotfi/gitdpeloy
	Url string
	// directory usually just created with ioutil.TempDir
	Directory string
	// PRIVATE token should be valid GitHub PAT, sourced from env var GITHUB_PAT
	token string
}

func main() {
	directory, err := tempFolder()
	if err != nil {
		log.Fatal(fmt.Errorf("Error getting tempdir: %w", err))
	}

	pat := os.Getenv("GITHUB_PAT")
	if len(pat) < 20 {
		log.Fatal("Github token not found")
	}

	repo := Repo{
		Url:       "https://github.com/maikdotfi/gitploy-dev",
		Directory: directory,
		token:     pat,
	}
	err = repo.cloneRepo()
	if err != nil {
		log.Fatal(err)
	}
	t := time.Now()
	err = repo.commit(fmt.Sprintf("Wow look at this! %s", t.Format("2006-01-02 15:04:05")))
	if err != nil {
		log.Fatal(err)
	}
	err = repo.push()
	if err != nil {
		log.Fatal(err)
	}
}

func tempFolder() (string, error) {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		return "", fmt.Errorf("Error creating tempdir: %w", err)
	}
	return path, err
}

func (r Repo) cloneRepo() error {
	// Clone the given repository to the given directory
	fmt.Printf("git clone %s %s\n", r.Url, r.Directory)

	rc, err := git.PlainClone(r.Directory, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "abc123", // yes, this can be anything except an empty string
			Password: r.token,
		},
		URL:      r.Url,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("Error cloning the repository %v: %w", r.Url, err)
	}

	// ... retrieving the branch being pointed by HEAD
	ref, err := rc.Head()
	if err != nil {
		return fmt.Errorf("Error cloning the repository %v: %w", r.Url, err)
	}
	// ... retrieving the commit object
	commit, err := rc.CommitObject(ref.Hash())
	if err != nil {
		return fmt.Errorf("Error cloning the repository %v: %w", r.Url, err)
	}

	fmt.Println(commit)
	return nil
}

func (r Repo) commit(msg string) error {
	// Opens an already existing repository.
	rc, err := git.PlainOpen(r.Directory)
	if err != nil {
		return fmt.Errorf("Error opening the repo in directory %v: %w", r.Directory, err)
	}

	w, err := rc.Worktree()
	if err != nil {
		return fmt.Errorf("Error to get the Worktree: %w", err)
	}

	// ... we need a file to commit so let's create a new file inside of the
	// worktree of the project using the go standard library.
	fmt.Println("echo \"hello world!\" > example-git-file")
	filename := filepath.Join(r.Directory, "example-git-file")
	err = ioutil.WriteFile(filename, []byte(msg), 0644)
	if err != nil {
		return fmt.Errorf("Error writing to file %v: %w", filename, err)
	}

	// Adds the new file to the staging area.
	fmt.Println("git add example-git-file")
	_, err = w.Add("example-git-file")
	if err != nil {
		return fmt.Errorf("Error adding file %v: %w", filename, err)
	}

	// We can verify the current status of the worktree using the method Status.
	fmt.Println("git status --porcelain")
	status, err := w.Status()
	if err != nil {
		return fmt.Errorf("Error getting status file %w", err)
	}

	fmt.Println(status)

	// Commits the current staging area to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit Since version 5.0.1, we can omit the Author signature, being read
	// from the git config files.
	fmt.Println("git commit -m \"example go-git commit\"")
	commit, err := w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Git Ploy",
			Email: "gitploy@maik.fi",
			When:  time.Now(),
		},
	})

	if err != nil {
		return fmt.Errorf("Error commiting %w", err)
	}

	// Prints the current HEAD to verify that all worked well.
	fmt.Println("git show -s")
	obj, err := rc.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("Error git show %w", err)
	}

	fmt.Println(obj)
	return nil
}

func (r Repo) push() error {
	rc, err := git.PlainOpen(r.Directory)
	if err != nil {
		return fmt.Errorf("Error opening the repo in directory %v: %w", r.Directory, err)
	}

	fmt.Println("git push")
	// push using default options
	err = rc.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "abc123",
			Password: r.token,
		},
	})
	if err != nil {
		return fmt.Errorf("Error pushing commit to remote %v: %w", r.Url, err)
	}
	return nil
}
