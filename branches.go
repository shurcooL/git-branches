package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/shurcooL/go/pipeutil"
	"github.com/shurcooL/go/trim"
	"github.com/shurcooL/vcsstate"
	"golang.org/x/tools/go/vcs"
	"gopkg.in/pipe.v2"
)

const (
	iso8601  = "2006-01-02 15:04:05 -0700"
	twoWeeks = time.Hour * 24 * 14
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

// Branches returns a Markdown table of branches with ahead/behind information relative to master branch,
// for a git repository in dir.
func Branches(dir string, opt BranchesOptions) (string, error) {
	opt.fillMissing()

	staleBranches := 0

	vcs, err := vcsstate.NewVCS(vcs.ByCmd("git"))
	if err != nil {
		return "", err
	}
	localBranch, err := vcs.Branch(dir)
	if err != nil {
		return "", err
	}

	// line is tab-separated local branch, commiter date.
	// E.g., "master\t2016-03-03 15:01:11 -0800".
	branchInfo := func(line []byte) []byte {
		branchDate := strings.Split(trim.LastNewline(string(line)), "\t")
		if len(branchDate) != 2 {
			return []byte("error: len(branchDate) != 2")
		}

		// Sort by dates, hide stale (>= 2 weeks) branches unless -all flag.
		if !*allFlag {
			date, err := time.Parse(iso8601, branchDate[1])
			if err != nil {
				log.Fatalln(err)
			}
			if time.Since(date) >= twoWeeks {
				staleBranches++
				return nil
			}
		}

		branch := branchDate[0]
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
			pipe.Exec("git", "for-each-ref", "--format=%(refname:short)\t%(committerdate:iso8601)", "refs/heads"),
			pipe.Replace(branchInfo),
		),
	)

	out, err := pipeutil.OutputDir(p, dir)
	if err != nil {
		return "", err
	}

	if staleBranches > 0 {
		out = append(out, []byte(fmt.Sprintf("\n(%v stale branches not shown.)\n", staleBranches))...)
	}

	return string(out), nil
}

// BranchesRemote returns a Markdown table of branches with ahead/behind information relative to remote,
// for a git repository in dir.
func BranchesRemote(dir string) (string, error) {
	staleBranches := 0

	vcs, err := vcsstate.NewVCS(vcs.ByCmd("git"))
	if err != nil {
		return "", err
	}
	localBranch, err := vcs.Branch(dir)
	if err != nil {
		return "", err
	}

	// line is tab-separated local branch, remote branch, commiter date.
	// E.g., "master\torigin/master\t2016-03-03 15:01:11 -0800".
	branchRemoteInfo := func(line []byte) []byte {
		branchRemoteDate := strings.Split(trim.LastNewline(string(line)), "\t")
		if len(branchRemoteDate) != 3 {
			return []byte("error: len(branchRemoteDate) != 3")
		}

		// Sort by dates, hide stale (>= 2 weeks) branches unless -all flag.
		if !*allFlag {
			date, err := time.Parse(iso8601, branchRemoteDate[2])
			if err != nil {
				log.Fatalln(err)
			}
			if time.Since(date) >= twoWeeks {
				staleBranches++
				return nil
			}
		}

		branch := branchRemoteDate[0]
		branchDisplay := branch
		if branch == localBranch {
			branchDisplay = "**" + branch + "**"
		}

		remote := branchRemoteDate[1]
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

	p := pipe.Script(
		pipe.Println("Branch | Remote | Behind | Ahead"),
		pipe.Println("-------|--------|-------:|:-----"),
		pipe.Line(
			pipe.Exec("git", "for-each-ref", "--format=%(refname:short)\t%(upstream:short)\t%(committerdate:iso8601)", "refs/heads"),
			pipe.Replace(branchRemoteInfo),
		),
	)

	out, err := pipeutil.OutputDir(p, dir)
	if err != nil {
		return "", err
	}

	if staleBranches > 0 {
		out = append(out, []byte(fmt.Sprintf("\n(%v stale branches not shown.)\n", staleBranches))...)
	}

	return string(out), nil
}
