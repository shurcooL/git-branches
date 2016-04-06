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
	baseFlag   = flag.String("base", "", "base branch to compare against (only when -remote is not specified)")
	remoteFlag = flag.Bool("remote", false, "compare local branches against remote")
	allFlag    = flag.Bool("all", false, "display all branches, including stale (>= 2 weeks old)")
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(2)
	}
	if *baseFlag != "" && *remoteFlag {
		fmt.Fprintln(os.Stderr, "warning: -base is ignored when -remote is specified")
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

	if *remoteFlag {
		cmd := exec.Command("git", "remote", "update", "--prune")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintln(os.Stderr, "git remote update failed:", err)
			os.Stderr.Write(out)
		}
	}

	var branches string
	switch *remoteFlag {
	case false:
		branches, err = Branches(dir, BranchesOptions{Base: *baseFlag})
	case true:
		branches, err = BranchesRemote(dir)
	}
	if err != nil {
		return err
	}

	stdout := int(os.Stdout.Fd())
	formatted, err := markdown.Process("", []byte(branches), &markdown.Options{Terminal: terminal.IsTerminal(stdout)})
	if err != nil {
		return err
	}
	os.Stdout.Write(formatted)

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
