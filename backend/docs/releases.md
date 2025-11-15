# Releasing `kthulu-cli`

Kthulu Forge uses [GoReleaser](https://goreleaser.com) to produce cross-platform builds of the generator CLI and publish them as GitHub Releases. This workflow is already wired inside the repository; you just need to tag a version and push it.

## Release architecture

- **Config** — `backend/backend/.goreleaser.yaml` defines how binaries are built, which ldflags inject version/build metadata, archive naming, checksums, and optional signing steps.
- **Automation** — `.github/workflows/release.yml` (inside `backend/`) runs on every `v*` tag. It checks out code, installs Go (using `backend/backend/go.mod`), runs the unit tests, imports the optional signing keys, and finally executes `goreleaser release --clean` from the backend root.
- **Artifacts** — Linux, macOS, and Windows binaries (amd64 + arm64) for `kthulu-cli` plus a checksum file are attached to the tag’s GitHub Release. When signing secrets are configured, `goreleaser` signs the checksum file so downstream automation can verify authenticity.

## First-time setup

1. **Configure repository secrets** (only needs to happen once):
   - `GORELEASER_GPG_PRIVATE_KEY` / `GORELEASER_GPG_PASSPHRASE` — optional but recommended so GoReleaser can sign the checksum manifest. Export your signing key in ASCII-armored form before pasting it as the secret value.
   - `COSIGN_PRIVATE_KEY` / `COSIGN_PASSWORD` — optional. If present, the workflow will make `COSIGN_PRIVATE_KEY_FILE` available so you can extend the pipeline with container or binary signing.
2. **Install GoReleaser locally** so you can run dry runs before tagging:
   ```sh
   curl -sSfL https://raw.githubusercontent.com/goreleaser/goreleaser/main/scripts/install.sh | sh -s -- -b $(go env GOPATH)/bin
   ```

## Preparing a release

1. **Run tests** from the backend root to make sure the tree is green.
   ```sh
   cd backend/backend
   go test ./...
   ```
2. **Check the changelog** (or update docs) so you can summarize what’s shipping.
3. **Verify the release locally** with a snapshot build. This exercise catches packaging issues without publishing anything:
   ```sh
   cd backend/backend
   goreleaser release --snapshot --skip-publish --clean
   ```
   Snapshot archives land under `dist/`; inspect one to confirm the CLI launches.
4. **Tag the release** using the `vMAJOR.MINOR.PATCH` scheme and push it:
   ```sh
   git tag -a v0.1.0 -m "kthulu-cli v0.1.0"
   git push origin v0.1.0
   ```
5. **Monitor the GitHub Action** (`Release`) triggered by the pushed tag. When it finishes you will see a draft release populated with all binaries and checksums. If you enabled signing, the checksum manifest will include the detached signature.

## Consuming the published binaries

Once the Release workflow completes:

- Download the appropriate archive for your platform from the GitHub Release page.
- Verify the checksum (and signature when available).
- Extract the `kthulu-cli` binary and place it on your `PATH`.

You can still distribute via `go install github.com/kthulu/kthulu-go/backend/cmd/kthulu-cli@<tag>` if you prefer building from source, but Release artifacts remain the canonical distribution for the generator CLI.
