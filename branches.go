package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/shurcooL/go/pipe_util"
	"github.com/shurcooL/go/trim"
	"gopkg.in/pipe.v2"
)

// BranchesOptions are options for Branches.
type BranchesOptions struct {
	Base string // Base branch to compare against (if blank, defaults to "master").
}

// fillMissing sets default values for mandatory options that are left empty.
func (opt *BranchesOptions) fillMissing() {
	if opt.Base == "" {
		opt.Base = "master"
	}
}

// Branches returns a Markdown table of branches with ahead/behind information relative to master branch.
func Branches(dir string, localBranch string, opt BranchesOptions) string {
	opt.fillMissing()
	branchInfo := func(line []byte) []byte {
		branch := trim.LastNewline(string(line))
		branchDisplay := branch
		if branch == localBranch {
			branchDisplay = "**" + branch + "**"
		}

		cmd := exec.Command("git", "rev-list", "--count", "--left-right", opt.Base+"..."+branch)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			log.Printf("error running %v: %v\n", cmd.Args, err)
			return []byte(fmt.Sprintf("%s | ? | ?\n", branchDisplay))
		}

		behindAhead := strings.Split(trim.LastNewline(string(out)), "\t")
		return []byte(fmt.Sprintf("%s | %s | %s\n", branchDisplay, behindAhead[0], behindAhead[1]))
	}

	p := pipe.Script(
		pipe.Println("Branch | Behind | Ahead"),
		pipe.Println("-------|-------:|:-----"),
		pipe.Line(
			pipe.Exec("git", "for-each-ref", "--format=%(refname:short)", "refs/heads"),
			pipe.Replace(branchInfo),
		),
	)

	out, err := pipe_util.OutputDir(p, dir)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// Input is a line containing tab-separated local branch and remote branch.
// For example, "master\torigin/master".
func branchRemoteInfo(dir string, localBranch string) func(line []byte) []byte {
	return func(line []byte) []byte {
		branchRemote := strings.Split(trim.LastNewline(string(line)), "\t")
		if len(branchRemote) != 2 {
			return []byte("error: len(branchRemote) != 2")
		}

		branch := branchRemote[0]
		branchDisplay := branch
		if branch == localBranch {
			branchDisplay = "**" + branch + "**"
		}

		remote := branchRemote[1]
		if remote == "" {
			return []byte(fmt.Sprintf("%s | | | \n", branchDisplay))
		}

		cmd := exec.Command("git", "rev-list", "--count", "--left-right", remote+"..."+branch)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			// This usually happens when the remote branch is gone.
			remoteDisplay := "~~" + remote + "~~"
			return []byte(fmt.Sprintf("%s | %s | | \n", branchDisplay, remoteDisplay))
		}

		behindAhead := strings.Split(trim.LastNewline(string(out)), "\t")
		return []byte(fmt.Sprintf("%s | %s | %s | %s\n", branchDisplay, remote, behindAhead[0], behindAhead[1]))
	}
}

// BranchesRemote returns a Markdown table of branches with ahead/behind information relative to remote.
func BranchesRemote(dir string, localBranch string) string {
	p := pipe.Script(
		pipe.Println("Branch | Remote | Behind | Ahead"),
		pipe.Println("-------|--------|-------:|:-----"),
		pipe.Line(
			pipe.Exec("git", "for-each-ref", "--format=%(refname:short)\t%(upstream:short)", "refs/heads"),
			pipe.Replace(branchRemoteInfo(dir, localBranch)),
		),
	)

	out, err := pipe_util.OutputDir(p, dir)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
