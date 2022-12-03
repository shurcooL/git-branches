git-branches
============

[![Go Reference](https://pkg.go.dev/badge/github.com/shurcooL/git-branches.svg)](https://pkg.go.dev/github.com/shurcooL/git-branches)

The git-branches command displays branches with behind/ahead commit counts. If you've used GitHub, you'll notice their branches page displays this info. This is that, but for your local git repo, in the terminal.

Go to any git repo you're working on and type:

```
$ git branch
  feature/sanitized_anchor_context
* main
  unfinished-attempt/alternative-sort
  wip/get-tweet-body
```

Now, run `go install github.com/shurcooL/git-branches@latest` on any machine with Go installed, and then:

```
$ git-branches
| Branch                              | Base | Behind | Ahead |
|-------------------------------------|------|-------:|:------|
| feature/sanitized_anchor_context    | main |    119 | 2     |
| **main**                            | main |      0 | 0     |
| unfinished-attempt/alternative-sort | main |     67 | 4     |
| wip/get-tweet-body                  | main |     40 | 1     |

| Branch                              | Remote                                     | Behind | Ahead |
|-------------------------------------|--------------------------------------------|-------:|:------|
| feature/sanitized_anchor_context    | origin/feature/sanitized_anchor_context    |      5 | 0     |
| **main**                            | origin/main                                |      0 | 1     |
| unfinished-attempt/alternative-sort | origin/unfinished-attempt/alternative-sort |      0 | 0     |
| wip/get-tweet-body                  |                                            |        |       |

(4 stale/trashed branches not shown.)
```

The currently checked out branch is emphasized with asterisks. You can see how branches compare to the base branch "main" locally. It supports -base option if you want to compare against a different base branch.

It's also easy to see which branches are up to date with remote, which ones need to be pushed/pulled.

Branches that are stale (>= 8 weeks old) or trashed ("trash/" prefix) are hidden by default, unless -all flag is used.

Installation
------------

```sh
go install github.com/shurcooL/git-branches@latest
```

License
-------

-	[MIT License](LICENSE)
