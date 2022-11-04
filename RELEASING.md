Releasing
=========

Setup
-----

Configure [gpg](https://www.gnupg.org/) so you can sign release artifacts.

Major and Minor Releases
------------------------

1.	Update the version numbers in the [install docs](docs/install.md) and commit to the main branch.

2.	Create a release branch:

	```
	export RELEASE_BRANCH=v$MAJOR.$MINOR.x
	git checkout -b $RELEASE_BRANCH
	git push origin $RELEASE_BRANCH
	```

3.	Tag the release and build artifacts:

	```
	./release.sh VERSION NAME
	```

	-	VERSION should have the form "1.2.3"
	-	NAME should be the name of the release ("Zeno", "Frege", "Heraclitus", etc.)

4.	Create the release in the Github UI.

	-	Edit the release notes
	-	Upload artifacts from the ./dist directory.

Patch Releases
--------------

Patch releases should be used to fix critical bugs (panics, data corruption, security vulnerabilities, etc.).

1.	Update the version numbers in the [install docs](docs/install.md) and commit to the main branch.

2.	Checkout the release branch:

	```
	export RELEASE_BRANCH=v$MAJOR.$MINOR.x
	git checkout $RELEASE_BRANCH
	```

3.	Cherry-pick commits onto the release branch.

	```
	git cherry-pick $SHA
	git push origin $RELEASE_BRANCH
	```

4.	Tag and publish the release (same as a major/minor release).
