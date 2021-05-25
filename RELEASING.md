Releasing
=========

Setup
-----

1.	Install [goreleaser](https://goreleaser.com):

	```
	go install github.com/goreleaser/goreleaser@latest
	```

2.	Configure a GitHub token with `repo` scope:

	```
	# https://github.com/settings/tokens/new
	export GITHUB_TOKEN="<TOKEN>"
	```

3.	Configure [gpg](https://www.gnupg.org/) so you can sign release artifacts.

Major and Minor Releases
------------------------

1.	Create a release branch:

	```
	RELEASE_BRANCH=v$MAJOR.$MINOR.x
	git checkout -b $RELEASE_BRANCH
	git push origin $RELEASE_BRANCH
	```

2.	Tag the release:

	```
	RELEASE_TAG=v$MAJOR.$MINOR.0
	RELEASE_NAME="<name>"
	git tag -s -a $RELEASE_TAG -m $RELEASE_NAME
	git push origin $RELEASE_TAG
	```

3.	Build and publish to GitHub:

	```
	goreleaser release
	```

4.	Find the [release in GitHub](https://github.com/aretext/aretext/releases/) and edit the release notes.

Patch Releases
--------------

Patch releases should be used to fix critical bugs (panics, data corruption, security vulnerabilities, etc.).

1.	Checkout the release branch:

	```
	RELEASE_BRANCH=v$MAJOR.$MINOR.x
	git checkout $RELEASE_BRANCH
	```

2.	Cherry-pick commits onto the release branch.

	```
	git cherry-pick $SHA
	git push origin $RELEASE_BRANCH
	```

3.	Tag and publish the release (same as a major/minor release).
