git-branches
============

[![Build Status](https://travis-ci.org/shurcooL/git-branches.svg?branch=master)](https://travis-ci.org/shurcooL/git-branches) [![GoDoc](https://godoc.org/github.com/shurcooL/git-branches?status.svg)](https://godoc.org/github.com/shurcooL/git-branches)

git-branches is a go gettable command that displays branches with behind/ahead commit counts. If you've used GitHub, you'll notice their branches page displays this info. This is that, but for your local git repo, in the terminal.

Go to any git repo you're working on and type:

```
$ git branch
  feature/sanitized_anchor_context
* master
  unfinished-attempt/alternative-sort
  wip/get-tweet-body
```

Now, run `go get -u github.com/shurcooL/git-branches` on any machine with Go installed, and then:

```
$ git-branches
| Branch                              | Base   | Behind | Ahead |
|-------------------------------------|--------|-------:|:------|
| feature/sanitized_anchor_context    | master |    119 | 2     |
| **master**                          | master |      0 | 0     |
| unfinished-attempt/alternative-sort | master |     67 | 4     |
| wip/get-tweet-body                  | master |     40 | 1     |

| Branch                              | Remote                                     | Behind | Ahead |
|-------------------------------------|--------------------------------------------|-------:|:------|
| feature/sanitized_anchor_context    | origin/feature/sanitized_anchor_context    |      5 | 0     |
| **master**                          | origin/master                              |      0 | 1     |
| unfinished-attempt/alternative-sort | origin/unfinished-attempt/alternative-sort |      0 | 0     |
| wip/get-tweet-body                  |                                            |        |       |

(4 stale/trashed branches not shown.)
```

The currently checked out branch is emphasized with asterisks. You can see how branches compare to master base branch locally. It supports -base option if you want to compare against a different base branch.

It's also easy to see which branches are up to date with remote, which ones need to be pushed/pulled.

Branches that are stale (>= 2 weeks old) or trashed ("trash/" prefix) are hidden by default, unless -all flag is used.

Installation
------------

```bash
go get -u github.com/shurcooL/git-branches
```

License
-------

-	[MIT License](https://opensource.org/licenses/mit-license.php)
