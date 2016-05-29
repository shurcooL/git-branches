// git-branches prints the commit behind/ahead counts for branches.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shurcooL/markdownfmt/markdown"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	baseFlag = flag.String("base", "", "base branch to compare against (only when -remote is not specified)")
	allFlag  = flag.Bool("all", false, "display all branches, including stale (>= 2 weeks old)")
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(2)
	}

	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	dir, err := fromDir(cwd)
	if err != nil {
		return err
	}
	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))

	// Display local branches.
	branches, staleBranches, err := Branches(dir, BranchesOptions{Base: *baseFlag})
	if err != nil {
		return err
	}
	formatted, err := markdown.Process("", []byte(branches), &markdown.Options{Terminal: isTerminal})
	if err != nil {
		return err
	}
	os.Stdout.Write(formatted)

	// Update remotes.
	cmd := exec.Command("git", "remote", "update", "--prune")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, "git remote update failed:", err)
		os.Stderr.Write(out)
	}

	fmt.Println()

	// Display remote branches.
	branches, staleRemoteBranches, err := BranchesRemote(dir)
	if err != nil {
		return err
	}
	formatted, err = markdown.Process("", []byte(branches), &markdown.Options{Terminal: isTerminal})
	if err != nil {
		return err
	}
	os.Stdout.Write(formatted)

	switch {
	case staleBranches == staleRemoteBranches && staleBranches > 0:
		fmt.Printf("\n(%v stale branches not shown.)\n", staleBranches)
	case staleBranches != staleRemoteBranches && (staleBranches > 0 || staleRemoteBranches > 0):
		fmt.Printf("\n(%v stale local, %v stale remote branches not shown.)\n", staleBranches, staleRemoteBranches)
	}

	return nil
}

// fromDir inspects dir and its parents to determine if it's inside a git repository.
// On return, root is the path corresponding to the root of the repository.
func fromDir(dir string) (root string, err error) {
	dir = filepath.Clean(dir)

	for {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return dir, nil
		}

		// Move to parent.
		ndir := filepath.Dir(dir)
		if len(ndir) >= len(dir) {
			break
		}
		dir = ndir
	}

	return "", fmt.Errorf("directory %q is not using git", dir)
}
