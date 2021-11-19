Contribution Guidelines
=======================

This document describes how you can contribute to the aretext project.

Ways to Contribute
------------------

1.	**Answer questions in the forum**: You can help other users by answering support questions in [GitHub Discussions](https://github.com/aretext/aretext/discussions). Please use a respectful tone.

2.	**Report bugs**: If you find a bug, please report it so we can fix it. Please open a [GitHub issue](https://github.com/aretext/aretext/issues) and fill out the bug report template. Provide as much detail as possible -- especially steps to reproduce the bug!

3.	**Maintain a package**: If you want to install aretext on a platform, but no one has created a package for it yet, consider adding one! You can help ensure that aretext is easily installable on as many platforms as possible.

4.	**Contribute code**: See below for instructions.

Security
--------

Please do NOT post security issues in public. To report a vulnerability, please send an email to [security@aretext.org](mailto:security@aretext.org).

Contributing Code
-----------------

Aretext is a minimalist text editor. To avoid adding unnecessary complexity, we accept mainly three kinds of changes:

1.	**Add syntax highlighting for a new language**: We want aretext to support as many programming languages as possible.

2.	**Implement a vim command**: The editor should (eventually) implement most of vim's normal, insert, and visual mode commands. We try to match vim's behavior as closely as possible, except when the additional complexity outweighs the benefits.

3.	**Fix a bug**: fix a reported bug and add test cases to prevent regressions.

If you have an idea for a feature, please start a [discussion in the forum](https://github.com/aretext/aretext/discussions) or open a [GitHub issue](https://github.com/aretext/aretext/issues) and fill out the feature request template.

### Code Guidelines

1.	[Add tests](https://golang.org/pkg/testing/) for any new features or bug fixes.
2.	Run `make` before committing code. This will run `go generate`, `goimports`, and `go test`. Generated code should be checked into the repository.
3.	Follow this guide for writing commit messages: [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)
4.	Add a "Signed-off-by" trailer in the commit message to record your agreement with the [Developer Certificate of Origin](https://developercertificate.org/). Git will add the trailer automatically if you pass the `-s` flag to `git commit`.

### Submitting a Pull Request

1.	Before writing any code, please open a [GitHub issue](https://github.com/aretext/aretext/issues) and fill out one of the templates. This helps avoid duplicate work.

2.	Fork the repository, add your changes on a branch, then [submit a pull request](https://github.com/aretext/aretext/pulls). Please fill out the pull request template completely, and [allow edits from maintainers](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/allowing-changes-to-a-pull-request-branch-created-from-a-fork).

3.	Fix any failing tests.

4.	A maintainer will review your code as soon as possible. We are always busy, so please be patient with us.

5.	Make any requested changes by updating the branch in your fork. Use [`git rebase`](https://git-rebase.io/) to squash commits into a small number of logical, self-contained changes.
