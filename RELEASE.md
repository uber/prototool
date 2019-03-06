Release Process
===============

This document outlines how to create a release of prototool.

1.  Set up some environment variables for use later.

    ```
    # This is the version being released.
    VERSION=1.21.0

    # This is the branch from which $VERSION will be released.
    # This is almost always dev.
    BRANCH=dev
    ```

    ** If you are copying/pasting commands, make sure you actually set the right value for VERSION above. **

2.  Make sure you have the latest master.

    ```
    git checkout master
    git pull
    ```

3.  Merge the branch being released into master.

    ```
    git merge $BRANCH
    ```

4.  Alter the Unreleased entry in CHANGELOG.md to point to `$VERSION` and
    update the link at the bottom of the file. Use the format `YYYY-MM-DD` for
    the year.

    ```diff
    -## [Unreleased]
    +## [1.21.0] - 2017-10-23
    ```

    ```diff
    -[Unreleased]: https://github.com/uber/prototool/compare/v1.20.1...HEAD
    +[1.21.0]: https://github.com/uber/prototool/compare/v1.20.1...v1.21.0
    ```


5.  Update the version number in internal/vars/vars.go and the Installation
    section of the README.md and verify that it matches what is in the changelog.

    ```diff
    -const Version = "1.21.0-dev"
    +const Version = "1.21.0"
    ```

6.  Create a commit for the release.

    ```
    git add internal/vars/vars.go README.md CHANGELOG.md
    git commit -m "Preparing release v$VERSION"
    ```

7.  Tag and push the release.

    ```
    git tag -a "v$VERSION" -m "v$VERSION"
    git push origin master "v$VERSION"
    ```

8.  Go to <https://travis-ci.org/uber/prototool/builds> and cancel the
    build for `v$VERSION`.  If that Codecov build completes before the Codecov
    build for master, the code coverage for master will not get updated because
    only one branch gets updated per commit; this was verified with Codecov
    support. This will get tested by the build for master anyways.

9.  Build the release artifacts. This will put files in the release package
    that will be uploaded in the next step.

    ```
    make releasegen
    ls release
    ```

10. Go to <https://github.com/uber/prototool/tags> and edit the release notes
    of the new tag.  Copy the changelog entries for this release in the
    release notes and set the name of the release to the version number
    (`v$VERSION`). Upload the release artifacts from the release directory.

11. Switch back to development.

    ```
    git checkout $BRANCH
    git merge master
    ```

12. Add a placeholder for the next version to CHANGELOG.md and a new link at
    the bottom.

    ```diff
    +## [Unreleased]
    +- No changes yet.
    +
     ## [1.21.0] - 2017-10-23
    ```

    ```diff
    +[Unreleased]: https://github.com/uber/prototool/compare/v1.21.0...HEAD
     [1.21.0]: https://github.com/uber/prototool/compare/v1.20.1...v1.21.0
    ```

13. Update the version number in internal/vars/vars.go to a new minor version
    suffixed with `"-dev"`.

    ```diff
    -const Version = "1.21.0"
    +const Version = "1.22.0-dev"
    ```

14. Commit and push your changes.

    ```
    git add CHANGELOG.md internal/vars/vars.go
    git commit -m 'Back to development'
    git push origin $BRANCH
    ```

15. Update the Homebrew formula using `brew bump-formula-pr`. This will create
    a fork of github.com/Homebrew/homebrew-core and create a PR with the
    updated formula for Prototool.

    ```
    brew bump-formula-pr --url=https://github.com/uber/prototool/archive/v$VERSION.tar.gz prototool
    ```
