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
	iso8601Strict = time.RFC3339 // Can use time.RFC3339 to parse git's ISO8601 strict time format.
	twoWeeks      = time.Hour * 24 * 14
)

// branches returns a Markdown table of branches with ahead/behind information relative to master branch,
// for a git repository in dir. baseBranch is base branch to compare against, and is never hidden as stale.
func branches(dir string, baseBranch string) (_ string, staleBranches int, _ error) {
	git, err := vcsstate.NewVCS(vcs.ByCmd("git"))
	if err != nil {
		return "", 0, err
	}
	localBranch, err := git.Branch(dir)
	if err != nil {
		return "", 0, err
	}
	// In general, we can't figure out what the "default" branch is, because we don't know which
	// remote, if any, is canonical. So just use the provided base branch.

	// line is tab-separated local branch, commiter date.
	// E.g., "master\t2016-03-03 15:01:11 -0800".
	branchInfo := func(line []byte) []byte {
		branchDate := strings.Split(trim.LastNewline(string(line)), "\t")
		if len(branchDate) != 2 {
			return []byte("error: len(branchDate) != 2")
		}

		branch := branchDate[0]
		branchDisplay := branch
		if branch == localBranch {
			branchDisplay = "**" + branch + "**"
		}

		// Hide stale (>= 2 weeks) and trashed ("trash/" prefix) branches,
		// unless -all flag or currently checked out or base branch.
		if !*allFlag && branch != localBranch && branch != baseBranch {
			if strings.HasPrefix(branch, "trash/") {
				staleBranches++
				return nil
			}

			date, err := time.Parse(iso8601Strict, branchDate[1])
			if err != nil {
				log.Fatalln(err)
			}
			if time.Since(date) >= twoWeeks {
				staleBranches++
				return nil
			}
		}

		cmd := exec.Command("git", "rev-list", "--count", "--left-right", baseBranch+"..."+branch)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			log.Printf("error running %v: %v\n", cmd.Args, err)
			return []byte(fmt.Sprintf("%s | %s | ? | ?\n", branchDisplay, baseBranch))
		}

		behindAhead := strings.Split(trim.LastNewline(string(out)), "\t")
		return []byte(fmt.Sprintf("%s | %s | %s | %s\n", branchDisplay, baseBranch, behindAhead[0], behindAhead[1]))
	}

	p := pipe.Script(
		pipe.Println("Branch | Base | Behind | Ahead"),
		pipe.Println("-------|------|-------:|:-----"),
		pipe.Line(
			pipe.Exec("git", "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)\t%(committerdate:iso8601-strict)", "refs/heads"),
			pipe.Replace(branchInfo),
		),
	)

	out, err := pipeutil.OutputDir(p, dir)
	if err != nil {
		return "", 0, err
	}

	return string(out), staleBranches, nil
}

// branchesRemote returns a Markdown table of branches with ahead/behind information relative to remote,
// for a git repository in dir. baseBranch is never hidden as stale.
func branchesRemote(dir string, baseBranch string) (_ string, staleBranches int, _ error) {
	git, err := vcsstate.NewVCS(vcs.ByCmd("git"))
	if err != nil {
		return "", 0, err
	}
	localBranch, err := git.Branch(dir)
	if err != nil {
		return "", 0, err
	}
	// In general, we can't figure out what the "default" branch is, because we don't know which
	// remote, if any, is canonical. So just use the provided base branch.

	// line is tab-separated local branch, remote branch, commiter date.
	// E.g., "master\torigin/master\t2016-03-03 15:01:11 -0800".
	branchRemoteInfo := func(line []byte) []byte {
		branchRemoteDate := strings.Split(trim.LastNewline(string(line)), "\t")
		if len(branchRemoteDate) != 3 {
			return []byte("error: len(branchRemoteDate) != 3")
		}

		branch := branchRemoteDate[0]
		branchDisplay := branch
		if branch == localBranch {
			branchDisplay = "**" + branch + "**"
		}

		// Hide stale (>= 2 weeks) and trashed ("trash/" prefix) branches,
		// unless -all flag or currently checked out or base branch.
		if !*allFlag && branch != localBranch && branch != baseBranch {
			if strings.HasPrefix(branch, "trash/") {
				staleBranches++
				return nil
			}

			date, err := time.Parse(iso8601Strict, branchRemoteDate[2])
			if err != nil {
				log.Fatalln(err)
			}
			if time.Since(date) >= twoWeeks {
				staleBranches++
				return nil
			}
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
			pipe.Exec("git", "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)\t%(upstream:short)\t%(committerdate:iso8601-strict)", "refs/heads"),
			pipe.Replace(branchRemoteInfo),
		),
	)

	out, err := pipeutil.OutputDir(p, dir)
	if err != nil {
		return "", 0, err
	}

	return string(out), staleBranches, nil
}
