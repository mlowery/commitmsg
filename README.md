# commitmsg

GitHub commit message assistance.

# Features

* Derives GitHub owner, repo, and issue from git branch name.
* Assuming the branch follows the naming convention, creates a commit message with the following:
    * GitHub issue title becomes the first line of the commit message. Its length is not altered although you may want to adhere to the conventional 50 character limit.
    * Third line of commit message is `Fixes #NNNN`. This triggers GitHub [auto-close behavior](https://help.github.com/articles/closing-issues-via-commit-messages/) on merge. This also triggers GitHub auto-linking between pull request and issue if you use [GitHub's `hub` tool](https://github.com/github/hub) to open pull requests since the default body of a pull request using `hub` is the third and later lines of the commit message.
    * GitHub issue body is placed in the commit message (only as a comment for reference).
* Respects any existing commit message content (such as `~/.gitmessage`).
* Does not attempt to connect to GitHub unless it is a new commit (not an amend) on a branch that follows the convention.

# Usage

Your branch name must follow this convention:

`<owner>/<repo>#<number>`

Example:

`owner/repo#1`

# Building and Installing

## One-time

```bash
$ docker run --rm -v "$PWD":/usr/src/commitmsg -w /usr/src/commitmsg -e GOOS=darwin -e GOARCH=amd64 golang:1.6 bash -c make
$ cp ./commitmsg <somewhere-on-path>
$ echo 'export COMMITMSG_GITHUB_BASE_URL=abc' >> ~/.bashrc  # e.g. https://api.github.com/
$ echo 'export COMMITMSG_ACCESS_TOKEN=xyz' >> ~/.bashrc
```

## Every clone

```bash
$ cat << 'EOF' > .git/hooks/prepare-commit-msg
#!/bin/sh

commitmsg "${@}"
EOF
$ chmod +x .git/hooks/prepare-commit-msg
```

# References

* [Closing issues via commit messages](https://help.github.com/articles/closing-issues-via-commit-messages/)
* [Autolinked references and URLs](https://help.github.com/articles/autolinked-references-and-urls/#issues-and-pull-requests)
* [hub helps you win at git.](https://github.com/github/hub)
* [Creating an access token for command-line use](https://help.github.com/articles/creating-an-access-token-for-command-line-use/)