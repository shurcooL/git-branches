// Command git-branches prints the commit behind/ahead counts for branches.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shurcooL/go/exp/12"
	"github.com/shurcooL/go/u/u6"
	legacyvcs "github.com/shurcooL/go/vcs"
	"github.com/shurcooL/markdownfmt/markdown"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	_ "sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
)

var (
	baseFlag   = flag.String("base", "", "base branch to compare against (only when -remote is not specified)")
	remoteFlag = flag.Bool("remote", false, "compare local branches against remote")
	oldFlag    = flag.Bool("old", true, "use old code path")
)

func main() {
	flag.Parse()
	if *baseFlag != "" && *remoteFlag {
		fmt.Fprintln(os.Stderr, "warning: -base is ignored when -remote is specified")
	}

	switch *oldFlag {
	case false:
		err := run()
		if err != nil {
			log.Fatalln(err)
		}
	case true:
		old()
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
	repo, err := vcs.Open("git", dir)
	if err != nil {
		return err
	}

	base := *baseFlag
	if base == "" {
		base = "master"
	}
	opt := vcs.BranchesOptions{
		BehindAheadBranch: base,
	}
	branches, err := repo.Branches(opt)
	if err != nil {
		return err
	}

	fmt.Printf("# Branches (%d total):\n", len(branches))
	for _, b := range branches {
		fmt.Printf("-%v | +%v | %s\n", b.Counts.Behind, b.Counts.Ahead, b.Name)
	}

	// TODO.
	return fmt.Errorf("not implemented")
}

// fromDir inspects dir and its parents to determine the
// version control system and code repository to use.
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

func old() {
	dir := exp12.LookupDirectory(".")
	dir.Update()
	if dir.Repo == nil {
		fmt.Fprintln(os.Stderr, "cwd has no git repository")
		os.Exit(1)
	}
	if dir.Repo.Vcs.Type() != legacyvcs.Git {
		fmt.Fprintln(os.Stderr, "non-git repos are not yet supported")
		os.Exit(1)
	}

	if *remoteFlag {
		cmd := exec.Command("git", "remote", "update", "--prune")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintln(os.Stderr, "git remote update failed:", err)
			os.Stderr.Write(out)
		}
	}
	dir.Repo.VcsLocal.Update()

	var branches string
	switch *remoteFlag {
	case false:
		branches = u6.Branches(dir.Repo, u6.BranchesOptions{Base: *baseFlag})
	case true:
		branches = u6.BranchesRemote(dir.Repo)
	}

	formatted, err := markdown.Process("", []byte(branches), nil)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(formatted)
}
