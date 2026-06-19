<div align="center">
  <a href="https://modtale.net">
    <img src="logo.svg" alt="Modtale Logo" width="850" height="132">
  </a>

  <h1 align="center">Modtale Publishing Examples</h1>

  <p align="center">
    <b>Reference publishing flows for the current Modtale API v1.</b><br>
    Create drafts, sync metadata, upload releases, and attach media from CI or local tooling.
  </p>

  <p align="center">
    <a href="https://modtale.net">
      <img src="https://img.shields.io/badge/Website-Modtale.net-3b82f6?style=for-the-badge" alt="Website" />
    </a>
    <a href="https://modtale.net/api-docs">
      <img src="https://img.shields.io/badge/API-Documentation-1e293b?style=for-the-badge&logo=swagger&logoColor=white" alt="Docs" />
    </a>
  </p>
</div>

---

## What This Covers

The examples target the live Modtale API v1 base URL:

```text
https://api.modtale.net/api/v1
```

Covered publishing features:

- Create a draft project with `POST /projects`
- Sync project metadata with `PUT /projects/{id}`
- Inspect plugin manifests for dependency suggestions with `POST /projects/{id}/versions/dependency-suggestions`
- Upload versions with `POST /projects/{id}/versions`
- Set optional version fields: `gameVersions`, `modIds`, `incompatibleProjectIds`, `channel`, `replaceExisting`, and `changelog`
- Upload icon, banner, gallery images, and YouTube gallery entries
- Publish or republish projects with `POST /projects/{id}/publish`
- Download artifacts through the current signed URL flow

Direct `publish` of a new project is an admin operation; project owners can use publish for eligible republish/restoration flows when they have the permission.

For raw request examples, see [`api-recipes.md`](api-recipes.md). That file is intentionally just copy-paste HTTP recipes, not another publisher program.

---

## Authentication

Create an API key from Modtale Developer Settings and pass it with:

```text
X-MODTALE-KEY: your_key
```

Most examples read either `MODTALE_KEY` or `MODTALE_API_KEY`.

Useful project-scoped permissions:

| Feature | Permission |
| --- | --- |
| Create a draft | `PROJECT_CREATE` |
| Read project/version metadata | `PROJECT_READ` |
| Update title, slug, description, tags, links, wiki, comments, carousel | `PROJECT_EDIT_METADATA` |
| Upload icon/banner | `PROJECT_EDIT_ICON`, `PROJECT_EDIT_BANNER` |
| Upload gallery image or YouTube gallery entry | `PROJECT_GALLERY_ADD` |
| Upload a version or inspect dependencies | `VERSION_CREATE` |
| Update an existing version's metadata | `VERSION_EDIT` |
| Revert, archive, unlist, publish | `PROJECT_STATUS_REVERT`, `PROJECT_STATUS_ARCHIVE`, `PROJECT_STATUS_UNLIST`, `PROJECT_STATUS_PUBLISH` |

---

## Environment

Common variables:

```bash
export MODTALE_KEY="md_your_key"
export MODTALE_PROJECT_ID="your-project-id"
export MODTALE_GAME_VERSIONS="2026.01.17-4b0f30090"
export MODTALE_CHANNEL="RELEASE" # RELEASE, BETA, or ALPHA
```

Optional variables:

```bash
export MODTALE_API_URL="https://api.modtale.net/api/v1"
export MODTALE_DEPENDENCIES="dependency-project-id:1.2.0:optional,embedded-project-id:3.0.0:embedded"
export MODTALE_INCOMPATIBLE_PROJECTS="incompatible-project-id"
export MODTALE_REPLACE_EXISTING="true"
```

Dependency entries use:

```text
projectId:versionNumber[:optional][:embedded]
```

The manifest suggestion endpoint returns this exact format as `dependencyEntry`.

---

## Go Publisher

Location: `go-publisher/`

This is the most complete example in the repository. It uses only the Go standard library.

Publish a version:

```bash
cd go-publisher
go run . \
  --file ../gradle-plugin/build/libs/gradle-plugin-1.0.0.jar \
  --version 1.0.0 \
  --game-version 2026.01.17-4b0f30090 \
  --channel RELEASE \
  --changelog "Initial Modtale release"
```

Publish with optional version metadata:

```bash
go run . \
  --file build/libs/example.jar \
  --version 1.1.0 \
  --channel BETA \
  --dependency dependency-project-id:2.0.0:optional \
  --incompatible-project old-project-id \
  --replace-existing
```

Inspect a plugin manifest and apply suggested dependencies:

```bash
go run . \
  --file build/libs/example.jar \
  --suggest-dependencies \
  --use-suggested-dependencies \
  --version 1.1.0
```

Create a draft, sync metadata, upload media, then publish a version:

```bash
go run . \
  --create-project \
  --title "Example Plugin" \
  --classification PLUGIN \
  --slug example-plugin \
  --summary "A short summary for search and project cards." \
  --sync-metadata \
  --about "## About\nLong-form Markdown description." \
  --tag utilities \
  --link source=https://github.com/Modtale/modtale-example \
  --repository-url https://github.com/Modtale/modtale-example \
  --license MIT \
  --allow-comments=true \
  --allow-modpacks=true \
  --gallery-carousel-enabled=true \
  --icon assets/icon.png \
  --banner assets/banner.png \
  --gallery-image assets/screenshot.png \
  --gallery-youtube https://www.youtube.com/watch?v=dQw4w9WgXcQ \
  --file build/libs/example.jar \
  --version 1.0.0
```

Use `--dry-run` to print requests without sending them.

---

## GitHub Actions

Locations:

- `.github/workflows/publish-java-plugin.yml`
- `.github/workflows/publish-asset-pack.yml`

Repository secret:

```text
MODTALE_API_KEY
```

Recommended repository variables:

```text
MODTALE_PROJECT_ID
MODTALE_GAME_VERSIONS
MODTALE_CHANNEL
MODTALE_DEPENDENCIES
MODTALE_INCOMPATIBLE_PROJECTS
MODTALE_REPLACE_EXISTING
```

Both workflows run on GitHub Releases and also support manual `workflow_dispatch`. The Java workflow builds `gradle-plugin/`; the asset workflow zips `asset-pack/`.

---

## Gradle Publishing

Location: `gradle-plugin/`

Build and publish locally:

```bash
cd gradle-plugin
export MODTALE_KEY="md_your_key"
export MODTALE_PROJECT_ID="your-project-id"
export MODTALE_GAME_VERSIONS="2026.01.17-4b0f30090"
export MODTALE_CHANNEL="RELEASE"
gradle publishToModtale
```

Optional fields:

```bash
export MODTALE_DEPENDENCIES="dependency-project-id:1.2.0:optional"
export MODTALE_INCOMPATIBLE_PROJECTS="old-project-id"
export MODTALE_REPLACE_EXISTING="true"
```

---

## Maven Publishing

Location: `maven-plugin/`

```bash
cd maven-plugin
export MODTALE_KEY="md_your_key"
export MODTALE_PROJECT_ID="your-project-id"
mvn clean verify -Ppublish-modtale \
  -Dmodtale.key="$MODTALE_KEY" \
  -Dmodtale.gameVersions=2026.01.17-4b0f30090 \
  -Dmodtale.channel=RELEASE \
  -Dmodtale.replaceExisting=false \
  -Dmodtale.changelog="Uploaded via Maven"
```

The Maven profile intentionally stays minimal because XML is awkward for repeated multipart fields like dependencies. Use the Go publisher or GitHub Actions examples when you need `modIds` or `incompatibleProjectIds`.

---

## Consuming Downloads In Gradle

Location: `gradle-maven-dependency/`

The current API returns short-lived signed download URLs:

1. `GET /projects/{id}/versions/{version}/download-url`
2. Download the returned `/download/{token}` path before it expires

That means a stable Maven/Ivy artifact pattern is no longer the right model. This sample provides a `downloadModtaleArtifact` Gradle task that fetches the signed URL first, downloads the artifact into `build/modtale/`, and adds it as a local file dependency when `modtale.project` or `MODTALE_DEPENDENCY_PROJECT` is set.

```bash
cd gradle-maven-dependency
gradle build \
  -Pmodtale.project=levelingcore \
  -Pmodtale.version=0.9.11 \
  -Pmodtale.gameVersion=2026.01.17-4b0f30090
```

For private or unlisted projects, set `MODTALE_KEY` or `MODTALE_API_KEY` with `PROJECT_READ`.

---

## Project Layout

```text
.
├── .github/workflows/              # CI publishing examples
├── api-recipes.md                  # Raw curl examples for the publishing API
├── go-publisher/                   # Comprehensive standard-library CLI
├── gradle-plugin/                  # Local Gradle publish task
├── gradle-maven-dependency/        # Signed-download Gradle consumption example
├── maven-plugin/                   # Maven publish profile
└── asset-pack/                     # Folder zipped by the asset-pack workflow
```

---

Modtale is not affiliated with Hypixel Studios.
