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
    <a href="https://modtale.net/docs/api">
      <img src="https://img.shields.io/badge/API-Documentation-1e293b?style=for-the-badge&logo=swagger&logoColor=white" alt="Docs" />
    </a>
  </p>
</div>

---

## Prerequisites

Before integrating these examples, ensure you have the following ready:

1.  **A Modtale Account**
    * Sign in at [modtale.net](https://modtale.net).
2.  **An API Key**
    * Navigate to **Profile Icon** → **Developer Settings**.
    * Generate a new key (e.g., `md_abc123...`) and save it securely.
3.  **A Project ID**
    * Create a project manually on Modtale.
    * The ID is the UUID found at the end of your project URL.

---

## Recommended: GitHub Actions
We strongly recommend using **GitHub Actions** for continuous delivery. This ensures your project is automatically built and published to Modtale whenever you create a GitHub Release.

<table>
  <thead>
    <tr>
      <th width="50%" align="center">
        <h3>📦 Resource Packs</h3>
        <img src="https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=github-actions&logoColor=white" alt="GH Actions">
      </th>
      <th width="50%" align="center">
        <h3>☕ Java Plugins</h3>
        <img src="https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=github-actions&logoColor=white" alt="GH Actions">
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
          <li>Edit .github/workflows/publish-resource-pack.yml with your PROJECT_ID.</li>
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
          <li>Edit .github/workflows/publish-java-plugin.yml with your PROJECT_ID.</li>
          <li>Create a new Release (tag) to trigger.</li>
        </ol>
      </td>
    </tr>
  </tbody>
</table>

---

## Local Integration
Use these methods if you prefer to publish directly from your local terminal or IDE.

<table>
  <thead>
    <tr>
      <th width="50%" align="center">
        <h3>🐘 Gradle</h3>
        <img src="https://img.shields.io/badge/Gradle-02303A?style=for-the-badge&logo=Gradle&logoColor=white" alt="Gradle">
      </th>
      <th width="50%" align="center">
        <h3>🪶 Maven</h3>
        <img src="https://img.shields.io/badge/Maven-C71A36?style=for-the-badge&logo=apache-maven&logoColor=white" alt="Maven">
      </th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td valign="top">
        <p><b>Location:</b> /gradle-plugin</p>
        <p>Adds a custom publishToModtale task to your build script.</p>
        <br>
        <b>Setup:</b>
        <ul>
          <li>Open gradle-plugin/build.gradle</li>
          <li>Review the custom task logic.</li>
        </ul>
        <br>
        <b>Usage (Bash):</b>
        <pre lang="bash">
export MODTALE_KEY="your_key"
export MODTALE_PROJECT_ID="uuid"
./gradlew publishToModtale</pre>
        <b>Usage (PowerShell):</b>
        <pre lang="powershell">
$env:MODTALE_KEY="your_key"
$env:MODTALE_PROJECT_ID="uuid"
.\gradlew publishToModtale</pre>
      </td>
      <td valign="top">
        <p><b>Location:</b> /maven-plugin</p>
        <p>Uses the exec-maven-plugin within a specific build profile.</p>
        <br>
        <b>Setup:</b>
        <ul>
          <li>Open maven-plugin/pom.xml</li>
          <li>Review the &lt;profile&gt; configuration.</li>
        </ul>
        <br>
        <b>Usage (Bash/PowerShell):</b>
        <pre lang="bash">
export MODTALE_KEY="your_key"
export MODTALE_PROJECT_ID="uuid"

mvn clean package -Ppublish-modtale</pre>
</td>
</tr>
  </tbody>
</table>

<div align="center">
  <br>
  <p><i>Modtale is not affiliated with Hypixel Studios.</i></p>
</div>