// Command git-branches prints the commit behind/ahead counts for branches.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/shurcooL/go/exp/12"
	"github.com/shurcooL/go/u/u6"
	"github.com/shurcooL/go/vcs"
	"github.com/shurcooL/markdownfmt/markdown"
)

var baseFlag = flag.String("base", "", "base branch to compare against (only when -remote is not specified)")
var remoteFlag = flag.Bool("remote", false, "compare local branches against remote")

func main() {
	flag.Parse()
	if *baseFlag != "" && *remoteFlag {
		fmt.Fprintln(os.Stderr, "warning: -base is ignored when -remote is specified")
	}

	dir := exp12.LookupDirectory(".")
	dir.Update()
	if dir.Repo == nil {
		fmt.Fprintln(os.Stderr, "cwd has no git repository")
		os.Exit(1)
	}
	if dir.Repo.Vcs.Type() != vcs.Git {
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
