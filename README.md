<div align="center">
  <a href="https://modtale.net">
    <img src="mtale_f.svg" alt="Modtale Logo" width="850" height="132">
  </a>

  <h1 align="center">Modtale Publishing Examples</h1>

  <p align="center">
    <b>Automate your Hytale content distribution.</b><br>
    A collection of reference implementations for the Modtale Public Upload API.
  </p>

  <p align="center">
    <a href="https://modtale.net">
      <img src="https://img.shields.io/badge/Website-Modtale.net-3b82f6?style=for-the-badge&logo=hytale&logoColor=white" alt="Website" />
    </a>
    <a href="https://modtale.net/api-docs">
      <img src="https://img.shields.io/badge/API-Documentation-1e293b?style=for-the-badge&logo=swagger&logoColor=white" alt="Docs" />
    </a>
  </p>
</div>

---

## Prerequisites

Before integrating these examples, ensure you have the following ready:

1. **A Modtale Account**

   * Sign in at [modtale.net](https://modtale.net).

3. **An API Key**

   * Navigate to **Profile Icon → Developer Settings**

   * Generate a new key (e.g., `md_abc123...`)

5. **A Project ID**

   * Create a project manually on Modtale

   * The ID is the UUID found at the end of your project URL

---

## Recommended: GitHub Actions

We strongly recommend using **GitHub Actions** for continuous delivery. This ensures your project is automatically built and published to Modtale whenever you create a GitHub Release.

<table>
  <thead>
    <tr>
      <th width="50%" align="center">
        <h3>📦 Resource Packs</h3>
        <img src="https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=github-actions&logoColor=white">
      </th>
      <th width="50%" align="center">
        <h3>☕ Java Plugins</h3>
        <img src="https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=github-actions&logoColor=white">
      </th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td valign="top">
        <p><b>Location:</b> /resource-pack</p>
        <p>Best for texture packs, data packs, or pre-compiled assets. Zips the folder and uploads it automatically.</p>
        <br>
        <b>Setup:</b>
        <ol>
          <li>Go to Repo <b>Settings → Secrets → Actions</b>.</li>
          <li>Add secret: MODTALE_API_KEY.</li>
          <li>Edit <code>.github/workflows/publish-resource-pack.yml</code> with your PROJECT_ID.</li>
          <li>Create a new Release (tag) to trigger.</li>
        </ol>
      </td>
      <td valign="top">
        <p><b>Location:</b> /github-action-plugin</p>
        <p>Best for standard Mods/Plugins. Compiles the Java code using Gradle within the workflow, then uploads the resulting Jar.</p>
        <br>
        <b>Setup:</b>
        <ol>
          <li>Go to Repo <b>Settings → Secrets → Actions</b>.</li>
          <li>Add secret: MODTALE_API_KEY.</li>
          <li>Edit <code>.github/workflows/publish-java-plugin.yml</code> with your PROJECT_ID.</li>
          <li>Create a new Release (tag) to trigger.</li>
        </ol>
      </td>
    </tr>
  </tbody>
</table>

---

# Local Integration

Use these methods if you prefer to publish directly from your local terminal or IDE.

---

## 🐘 **Gradle**

### Location: `/gradle-plugin`

Adds a custom `publishToModtale` task to your build script.

**Setup:**

* Open `gradle-plugin/build.gradle`
* Review the custom task logic

**Usage (Bash):**

```bash
export MODTALE_KEY="your_key"
export MODTALE_PROJECT_ID="uuid"
./gradlew publishToModtale
```

**Usage (PowerShell):**

```powershell
$env:MODTALE_KEY="your_key"
$env:MODTALE_PROJECT_ID="uuid"
.\gradlew publishToModtale
```

---

## 🪶 **Maven**

### Location: `/maven-plugin`

Uses the `exec-maven-plugin` within a dedicated build profile.

**Setup:**

* Open `maven-plugin/pom.xml`
* Review the `<profile>` configuration

**Usage (Bash/PowerShell):**

```bash
export MODTALE_KEY="your_key"
export MODTALE_PROJECT_ID="uuid"

mvn clean package -Ppublish-modtale
```

---

## 🦫 **Go**

### Location: `/go-publisher`

A lightweight Go CLI that performs a multipart POST upload to Modtale.
This is ideal when you're automating deployments or building cross-platform tools.

**Setup:**

* Review `go-publisher/main.go`
* Ensure Go 1.21+ is installed

**Usage (Bash):**

```bash
export MODTALE_KEY="your_key"
export MODTALE_PROJECT_ID="uuid"
go run main.go --jar path/to/file.jar
```

**Usage (PowerShell):**

```powershell
$env:MODTALE_KEY="your_key"
$env:MODTALE_PROJECT_ID="uuid"
go run main.go --jar path\to\file.jar
```

If no `--jar` is provided, the CLI automatically searches for
`build/libs/*.jar`, matching the Java example below.

<div align="center">
  <br>
  <p><i>Modtale is not affiliated with Hypixel Studios.</i></p>
</div>

