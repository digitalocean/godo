# Contributing

We love contributions! You are welcome to open a pull request, but it's a good idea to
open an issue and discuss your idea with us first.

Once you are ready to open a PR, please keep the following guidelines in mind:

1. Code should be `go fmt` compliant.
1. Types, structs and funcs should be documented.
1. Tests pass.

## Getting set up

`godo` uses go modules. Just fork this repo, clone your fork and off you go!

## Client.V2 OpenAPI regeneration

`Client.V2` is generated from [digitalocean/openapi](https://github.com/digitalocean/openapi).

**Automated (preferred):** when OpenAPI merges to `main`, a bot workflow regenerates
`v2/*.gen.go` and opens a godo PR. You do **not** need to run `make generate` for
routine API updates.

1. In **godo**, workflow [`.github/workflows/v2-client-gen.yml`](.github/workflows/v2-client-gen.yml)
   downloads the published bundled spec, runs the generator, and opens a PR.
2. In **openapi**, merge
   `.github/workflows/trigger-godo-client-gen.yml`
   (same pattern as the pydo/dots triggers). It reuses secret
   `WORKFLOW_TRIGGER_TOKEN` (ensure the PAT can also dispatch workflows on
   `digitalocean/godo`).
3. Optional godo secret `WORKFLOW_TOKEN` — PAT for richer PR metadata; otherwise
   `GITHUB_TOKEN` is used.

**Manual / local:**

```sh
# sibling checkout of digitalocean/openapi
make generate

# or use a pre-bundled YAML (same path CI uses)
SPEC_FILE=/path/to/DigitalOcean-public.v2.yaml make generate
```

Pin file: `DO_OPENAPI_COMMIT_SHA.txt` records which OpenAPI commit the checked-in
`v2/*.gen.go` files came from.

## Running tests

When working on code in this repository, tests can be run via:

```sh
go test -mod=vendor .
```

## Versioning

Godo follows [semver](https://www.semver.org) versioning semantics.
New functionality should be accompanied by increment to the minor
version number. Any code merged to main is subject to release.

## Releasing

The repo uses GitHub workflows to publish a draft release when a new tag is
pushed. We use [semver](https://semver.org/#summary) to determine the version
number for the tag.

1. Run `make changes` to review the merged PRs since last release and decide what kind of release you are doing (bugfix, feature or breaking).
    * Review the tags on each PR and make sure they are categorized
      appropriately.

2. Run `BUMP=(bugfix|feature|breaking) make bump_version` to update the `godo`
   version.  
   `BUMP` also accepts `(patch|minor|major)`

   Command example:

   ```bash
   make BUMP=minor bump_version
   ```

3. Update the godo version in `godo.go` and add changelog generator logs in `CHANGELOG.md` file. Create a separate PR with only these changes.

4. Once the commit has been pushed, tag the commit to trigger the
   release workflow: run `make tag` to tag the latest commit and push the tag to ORIGIN.

   Notes:
   * To tag an earlier commit, run `COMMIT=${commit} make tag`.
   * To push the tag to a different remote, run `ORIGIN=${REMOTE} make tag`.

5. Once the release process completes, review the draft release for correctness and publish the release.  
   Ensure the release has been marked `Latest`.

6. Verify the module checksum against the Go checksum database (to catch retagging issues):

   1. Run `go mod tidy` inside doctl (or any other repo that uses godo) so that the corresponding godo hash in the `go.sum` file gets updated.
   2. Hit the curl command in the terminal for the recent release version, for example:

      ```bash
      curl https://sum.golang.org/lookup/github.com/digitalocean/godo@v1.165.1
      ```

   3. Both hashes should match for that specific godo version.

## Go Version Support

This project follows the support [policy of Go](https://go.dev/doc/devel/release#policy)
as its support policy. The two latest major releases of Go are supported by the project.
[CI workflows](.github/workflows/ci.yml) should test against both supported versions.
[go.mod](./go.mod) should specify the oldest of the supported versions to give
downstream users of godo flexibility.
