# Release procedure for Blubber

## Before you start

The script installs the tools that it needs. You must have:

 * The Go toolchain on your `PATH`.
 * A signing key, for `git tag --sign` to work.
 * Permission to push to the repository.

## Run the release script

Run `scripts/release.sh` which will:

 1. Increment the value in `VERSION` (minor by default; pass `-p` to do a
    patch release, or `-M` to do a major release).
 2. Generate `CHANGELOG.md` using `git chglog`.
 3. Show the changes and ask you to confirm them.
 4. Update the image references in `README.md` to the new tag.
 5. Commit `VERSION`, `CHANGELOG.md`, and `README.md`.
 6. Create a signed version tag.
 7. Open a merge request for the commit. The merge request merges when the
    pipeline succeeds.
 8. Push the version tag after the merge request merges.
