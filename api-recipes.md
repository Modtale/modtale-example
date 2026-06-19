# Modtale API Publishing Recipes

These examples show the raw HTTP requests behind the publisher examples. They are useful when wiring Modtale into an existing release job, CI system, or build tool without adopting another standalone program.

Set the common environment first:

```bash
export MODTALE_API_URL="https://api.modtale.net/api/v1"
export MODTALE_KEY="md_your_key"
export MODTALE_PROJECT_ID="your-project-id"
```

Use `--fail-with-body` in CI so failed API responses are visible in logs.

## Create A Draft

`POST /projects` creates a draft project and returns a project payload containing `id`.

```bash
curl --fail-with-body -X POST "$MODTALE_API_URL/projects" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -F "title=Example Plugin" \
  -F "classification=PLUGIN" \
  -F "description=A short summary for project cards and search." \
  -F "slug=example-plugin"
```

Valid classifications are `PLUGIN`, `DATA`, `ART`, `SAVE`, and `MODPACK`.

## Sync Project Metadata

Use `PUT /projects/{id}` when release automation should keep public project metadata aligned with the repository.

```bash
curl --fail-with-body -X PUT "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -H "Content-Type: application/json" \
  --data-binary @- <<'JSON'
{
  "title": "Example Plugin",
  "slug": "example-plugin",
  "description": "A short summary for project cards and search.",
  "about": "## Example Plugin\n\nLong-form Markdown project description.",
  "tags": ["utilities", "server"],
  "links": {
    "source": "https://github.com/Modtale/modtale-example",
    "issues": "https://github.com/Modtale/modtale-example/issues"
  },
  "repositoryUrl": "https://github.com/Modtale/modtale-example",
  "license": "MIT",
  "allowModpacks": true,
  "allowComments": true,
  "hmWikiEnabled": false,
  "galleryCarouselEnabled": true
}
JSON
```

Send only the fields your automation owns. Leaving a field out keeps the existing value.

## Inspect Manifest Dependencies

For plugin JARs, `POST /projects/{id}/versions/dependency-suggestions` can inspect the manifest and return suggested dependency entries. The response includes `suggestions[].dependencyEntry`, already formatted for version uploads.

```bash
curl --fail-with-body -X POST "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/versions/dependency-suggestions" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -F "file=@build/libs/example-plugin-1.2.0.jar"
```

Dependency entries use:

```text
projectId:versionNumber[:optional][:embedded]
```

Use `:optional` for dependencies players can skip. Use `:embedded` when the dependency is bundled into the artifact.

## Upload A Version

`POST /projects/{id}/versions` is multipart. Repeat list fields instead of joining them into one value when you can.

```bash
curl --fail-with-body -X POST "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/versions" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -F "file=@build/libs/example-plugin-1.2.0.jar" \
  -F "versionNumber=1.2.0" \
  -F "gameVersions=2026.01.17-4b0f30090" \
  -F "gameVersions=2026.01.13-dcad8778f" \
  -F "channel=BETA" \
  -F "modIds=dependency-project-id:2.0.0:optional" \
  -F "modIds=embedded-library-id:3.1.4:embedded" \
  -F "incompatibleProjectIds=old-conflicting-project-id" \
  -F "replaceExisting=true" \
  -F "changelog=Release notes in Markdown."
```

Channels are `RELEASE`, `BETA`, and `ALPHA`. `replaceExisting=true` is useful for rerunning a failed CI release for the same version target.

## Update Version Metadata

Use `PUT /projects/{id}/versions/{versionId}` when the artifact should stay the same but metadata needs correction.

```bash
curl --fail-with-body -X PUT "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/versions/$MODTALE_VERSION_ID" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -H "Content-Type: application/json" \
  --data-binary @- <<'JSON'
{
  "gameVersions": ["2026.01.17-4b0f30090"],
  "channel": "RELEASE",
  "modIds": ["dependency-project-id:2.0.0:optional"],
  "incompatibleProjectIds": ["old-conflicting-project-id"],
  "changelog": "Corrected release notes."
}
JSON
```

## Upload Media

Media endpoints are separate so release jobs can update only the assets they own.

```bash
curl --fail-with-body -X PUT "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/icon" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -F "file=@assets/icon.png"

curl --fail-with-body -X PUT "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/banner" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -F "file=@assets/banner.png"

curl --fail-with-body -X POST "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/gallery" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -F "file=@assets/screenshot.png"

curl --fail-with-body -X POST "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/gallery/youtube" \
  -H "X-MODTALE-KEY: $MODTALE_KEY" \
  -H "Content-Type: application/json" \
  --data '{"videoUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ"}'
```

## Change Visibility

Existing projects can be moved through lifecycle endpoints when the key has the matching permission:

```bash
curl --fail-with-body -X POST "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/unlist" \
  -H "X-MODTALE-KEY: $MODTALE_KEY"

curl --fail-with-body -X POST "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/private" \
  -H "X-MODTALE-KEY: $MODTALE_KEY"

curl --fail-with-body -X POST "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/archive" \
  -H "X-MODTALE-KEY: $MODTALE_KEY"
```

Direct `publish` of a new project is an admin action. For owner automation, use it only for eligible republish/restoration flows.

## Download A Version

Downloads use short-lived signed URLs.

```bash
curl --fail-with-body "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/versions/1.2.0/download-url?gameVersion=2026.01.17-4b0f30090" \
  -H "X-MODTALE-KEY: $MODTALE_KEY"
```

The response contains:

```json
{
  "downloadUrl": "/download/token-value",
  "expiresIn": 300
}
```

Fetch the returned path before it expires:

```bash
curl --fail-with-body -L "$MODTALE_API_URL/download/token-value" \
  -o example-plugin-1.2.0.jar
```

For dependency bundles, request a bundle URL and repeat `deps` for selected dependency project IDs:

```bash
curl --fail-with-body "$MODTALE_API_URL/projects/$MODTALE_PROJECT_ID/versions/1.2.0/download-bundle-url?gameVersion=2026.01.17-4b0f30090&deps=dependency-project-id" \
  -H "X-MODTALE-KEY: $MODTALE_KEY"
```
