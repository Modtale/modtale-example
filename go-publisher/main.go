package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultGameVersion = "2026.01.17-4b0f30090"

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*s = append(*s, part)
		}
	}
	return nil
}

type keyValueMap map[string]string

func (m *keyValueMap) String() string {
	if m == nil {
		return ""
	}
	keys := make([]string, 0, len(*m))
	for key := range *m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+(*m)[key])
	}
	return strings.Join(parts, ",")
}

func (m *keyValueMap) Set(value string) error {
	if *m == nil {
		*m = make(map[string]string)
	}

	key, val, ok := strings.Cut(value, "=")
	if !ok || strings.TrimSpace(key) == "" {
		return fmt.Errorf("expected key=value, got %q", value)
	}
	(*m)[strings.TrimSpace(key)] = strings.TrimSpace(val)
	return nil
}

type optionalBool struct {
	set   bool
	value bool
}

func (b *optionalBool) String() string {
	if !b.set {
		return ""
	}
	return strconv.FormatBool(b.value)
}

func (b *optionalBool) Set(value string) error {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	b.set = true
	b.value = parsed
	return nil
}

func (b *optionalBool) IsBoolFlag() bool {
	return true
}

type projectResponse struct {
	ID string `json:"id"`
}

type manifestInspectionResult struct {
	GameVersion string                         `json:"gameVersion"`
	ModVersion  string                         `json:"modVersion"`
	Suggestions []manifestDependencySuggestion `json:"suggestions"`
}

type manifestDependencySuggestion struct {
	DependencyEntry string `json:"dependencyEntry"`
}

type modtaleClient struct {
	apiBase      string
	projectsBase string
	apiKey       string
	dryRun       bool
	httpClient   *http.Client
}

func main() {
	log.SetFlags(0)

	var gameVersions stringList
	var dependencies stringList
	var incompatibleProjects stringList
	var tags stringList
	var links keyValueMap
	var galleryImages stringList
	var galleryVideos stringList

	var allowModpacks optionalBool
	var allowComments optionalBool
	var hmWikiEnabled optionalBool
	var galleryCarouselEnabled optionalBool

	baseURL := flag.String("base-url", envDefault("MODTALE_API_URL", "https://api.modtale.net/api/v1"), "Modtale API v1 base URL")
	legacyEndpoint := flag.String("endpoint", "", "Deprecated alias for the projects endpoint, for example https://api.modtale.net/api/v1/projects")
	apiKey := flag.String("api-key", envDefault("MODTALE_KEY", envDefault("MODTALE_API_KEY", "")), "Modtale API key; defaults to MODTALE_KEY or MODTALE_API_KEY")
	projectIDFlag := flag.String("project-id", envDefault("MODTALE_PROJECT_ID", ""), "Project ID or route key; defaults to MODTALE_PROJECT_ID")

	filePath := flag.String("file", "", "Artifact to upload, usually a .jar or .zip")
	jarPath := flag.String("jar", "", "Deprecated alias for --file")
	version := flag.String("version", envDefault("MODTALE_VERSION", ""), "Version number to publish")
	changelog := flag.String("changelog", envDefault("MODTALE_CHANGELOG", "Uploaded via Modtale Go publisher"), "Markdown changelog for the version")
	channel := flag.String("channel", envDefault("MODTALE_CHANNEL", "RELEASE"), "Release channel: RELEASE, BETA, or ALPHA")
	replaceExisting := flag.Bool("replace-existing", envBool("MODTALE_REPLACE_EXISTING", false), "Replace existing versions with the same version/game-version target")
	skipUpload := flag.Bool("skip-upload", false, "Skip the version upload; useful with --sync-metadata or --version-id")

	createProject := flag.Bool("create-project", false, "Create a new draft project before uploading")
	title := flag.String("title", "", "Project title for --create-project or --sync-metadata")
	classification := flag.String("classification", "PLUGIN", "Project classification for --create-project: PLUGIN, DATA, ART, SAVE, or MODPACK")
	summary := flag.String("summary", "", "Project short summary")
	slug := flag.String("slug", "", "Optional project slug")
	owner := flag.String("owner", "", "Optional owner user or organization ID when creating a draft")

	syncMetadata := flag.Bool("sync-metadata", false, "Update project metadata before uploading")
	about := flag.String("about", "", "Full Markdown project description")
	repositoryURL := flag.String("repository-url", "", "Repository URL, for example https://github.com/owner/repo")
	license := flag.String("license", "", "Project license identifier or name")
	hmWikiSlug := flag.String("hm-wiki-slug", "", "Optional Hytale Wiki slug")
	flag.Var(&allowModpacks, "allow-modpacks", "Set whether this project can be included in modpacks")
	flag.Var(&allowComments, "allow-comments", "Set whether comments are enabled")
	flag.Var(&hmWikiEnabled, "hm-wiki-enabled", "Set whether Hytale Wiki integration is enabled")
	flag.Var(&galleryCarouselEnabled, "gallery-carousel-enabled", "Set whether the gallery carousel is enabled")

	iconPath := flag.String("icon", "", "Optional icon image to upload")
	bannerPath := flag.String("banner", "", "Optional banner image to upload")
	versionID := flag.String("version-id", "", "Existing version ID to update with JSON metadata")
	suggestDependencies := flag.Bool("suggest-dependencies", false, "Inspect the artifact manifest and print dependency suggestions before uploading")
	useSuggestedDependencies := flag.Bool("use-suggested-dependencies", false, "Append manifest dependency suggestions to --dependency values")
	publish := flag.Bool("publish", false, "Publish or republish the project after uploading; new projects require admin approval")
	dryRun := flag.Bool("dry-run", false, "Print the planned requests without sending them")

	flag.Var(&gameVersions, "game-version", "Game version to target; repeat or pass comma-separated values")
	flag.Var(&gameVersions, "gameVersions", "Deprecated alias for --game-version")
	flag.Var(&dependencies, "dependency", "Dependency entry projectId:version[:optional][:embedded]; repeat or pass comma-separated values")
	flag.Var(&dependencies, "mod-id", "Alias for --dependency")
	flag.Var(&dependencies, "modIds", "Deprecated alias for --dependency")
	flag.Var(&incompatibleProjects, "incompatible-project", "Incompatible project ID; repeat or pass comma-separated values")
	flag.Var(&incompatibleProjects, "incompatibleProjectIds", "Deprecated alias for --incompatible-project")
	flag.Var(&tags, "tag", "Project tag for --sync-metadata; repeat or pass comma-separated values")
	flag.Var(&links, "link", "Project link as label=url; repeat for multiple links")
	flag.Var(&galleryImages, "gallery-image", "Gallery image path to upload; repeat for multiple images")
	flag.Var(&galleryVideos, "gallery-youtube", "YouTube URL to add to the gallery; repeat for multiple videos")
	flag.Parse()

	providedFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		providedFlags[f.Name] = true
	})
	updateChangelog := *changelog
	if !providedFlags["changelog"] && strings.TrimSpace(os.Getenv("MODTALE_CHANGELOG")) == "" {
		updateChangelog = ""
	}
	updateChannel := *channel
	if !providedFlags["channel"] && strings.TrimSpace(os.Getenv("MODTALE_CHANNEL")) == "" {
		updateChannel = ""
	}

	if *jarPath != "" && *filePath == "" {
		*filePath = *jarPath
	}
	if len(gameVersions) == 0 {
		_ = gameVersions.Set(envDefault("MODTALE_GAME_VERSIONS", defaultGameVersion))
	}
	if len(dependencies) == 0 {
		_ = dependencies.Set(os.Getenv("MODTALE_DEPENDENCIES"))
	}
	if len(incompatibleProjects) == 0 {
		_ = incompatibleProjects.Set(os.Getenv("MODTALE_INCOMPATIBLE_PROJECTS"))
	}
	updateGameVersions := []string(gameVersions)
	if !providedFlags["game-version"] && !providedFlags["gameVersions"] && strings.TrimSpace(os.Getenv("MODTALE_GAME_VERSIONS")) == "" {
		updateGameVersions = nil
	}

	apiBase, projectsBase := normalizeAPIEndpoints(*baseURL, *legacyEndpoint)
	client := &modtaleClient{
		apiBase:      apiBase,
		projectsBase: projectsBase,
		apiKey:       strings.TrimSpace(*apiKey),
		dryRun:       *dryRun,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}

	projectID := strings.TrimSpace(*projectIDFlag)
	if *createProject {
		requireAPIKey(client)
		createdID, err := client.createProject(*title, *classification, *summary, *slug, *owner)
		must(err)
		if createdID != "" {
			projectID = createdID
			fmt.Printf("Created draft project: %s\n", projectID)
		} else if client.dryRun && projectID == "" {
			projectID = "dry-run-project-id"
		}
	}
	if projectID == "" {
		log.Fatalf("MODTALE_PROJECT_ID or --project-id is required unless --create-project returns a project")
	}

	if *syncMetadata || metadataHasFields(*title, *summary, *slug, *about, *repositoryURL, *license, *hmWikiSlug, tags, links, allowModpacks, allowComments, hmWikiEnabled, galleryCarouselEnabled) {
		requireAPIKey(client)
		must(client.updateProjectMetadata(projectID, projectMetadataInput{
			title:                  *title,
			slug:                   *slug,
			description:            *summary,
			about:                  *about,
			tags:                   tags,
			links:                  links,
			repositoryURL:          *repositoryURL,
			license:                *license,
			allowModpacks:          allowModpacks,
			allowComments:          allowComments,
			hmWikiEnabled:          hmWikiEnabled,
			hmWikiSlug:             *hmWikiSlug,
			galleryCarouselEnabled: galleryCarouselEnabled,
		}))
		fmt.Println("Project metadata synced.")
	}

	artifact := strings.TrimSpace(*filePath)
	if artifact == "" && (!*skipUpload || *suggestDependencies) {
		var err error
		artifact, err = findDefaultArtifact()
		must(err)
	}

	if *suggestDependencies {
		requireAPIKey(client)
		result, err := client.inspectManifest(projectID, artifact)
		must(err)
		if result != nil {
			printManifestSuggestions(*result)
			if *useSuggestedDependencies {
				for _, suggestion := range result.Suggestions {
					if suggestion.DependencyEntry != "" {
						dependencies = append(dependencies, suggestion.DependencyEntry)
					}
				}
				if *version == "" && result.ModVersion != "" {
					*version = result.ModVersion
				}
				if result.GameVersion != "" && len(gameVersions) == 0 {
					gameVersions = append(gameVersions, result.GameVersion)
				}
			}
		}
	}

	if !*skipUpload {
		requireAPIKey(client)
		if artifact == "" {
			log.Fatalf("--file is required when upload is enabled")
		}
		actualVersion := strings.TrimSpace(*version)
		if actualVersion == "" {
			actualVersion = versionFromArtifact(artifact)
		}
		must(client.uploadVersion(projectID, artifact, actualVersion, gameVersions, dependencies, incompatibleProjects, *changelog, *channel, *replaceExisting))
		fmt.Printf("Uploaded version %s for project %s.\n", actualVersion, projectID)
	}

	if *versionID != "" {
		requireAPIKey(client)
		must(client.updateVersion(projectID, *versionID, updateGameVersions, dependencies, incompatibleProjects, updateChangelog, updateChannel))
		fmt.Printf("Updated version metadata for %s.\n", *versionID)
	}

	if *iconPath != "" {
		requireAPIKey(client)
		must(client.uploadProjectFile(projectID, "PUT", "icon", *iconPath))
		fmt.Println("Project icon uploaded.")
	}
	if *bannerPath != "" {
		requireAPIKey(client)
		must(client.uploadProjectFile(projectID, "PUT", "banner", *bannerPath))
		fmt.Println("Project banner uploaded.")
	}
	for _, image := range galleryImages {
		requireAPIKey(client)
		must(client.uploadProjectFile(projectID, "POST", "gallery", image))
		fmt.Printf("Gallery image uploaded: %s\n", image)
	}
	for _, videoURL := range galleryVideos {
		requireAPIKey(client)
		must(client.addGalleryVideo(projectID, videoURL))
		fmt.Printf("Gallery video added: %s\n", videoURL)
	}

	if *publish {
		requireAPIKey(client)
		must(client.projectTransition(projectID, "publish"))
		fmt.Println("Project publish request sent.")
	}
}

type projectMetadataInput struct {
	title                  string
	slug                   string
	description            string
	about                  string
	tags                   []string
	links                  map[string]string
	repositoryURL          string
	license                string
	allowModpacks          optionalBool
	allowComments          optionalBool
	hmWikiEnabled          optionalBool
	hmWikiSlug             string
	galleryCarouselEnabled optionalBool
}

func (c *modtaleClient) createProject(title string, classification string, summary string, slug string, owner string) (string, error) {
	if strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("--title is required with --create-project")
	}
	if strings.TrimSpace(classification) == "" {
		return "", fmt.Errorf("--classification is required with --create-project")
	}

	values := url.Values{}
	values.Set("title", title)
	values.Set("classification", strings.ToUpper(classification))
	if summary != "" {
		values.Set("description", summary)
	}
	if slug != "" {
		values.Set("slug", slug)
	}
	if owner != "" {
		values.Set("owner", owner)
	}

	if c.dryRun {
		fmt.Printf("DRY RUN POST %s\n%s\n", c.projectsBase, values.Encode())
		return "", nil
	}

	req, err := http.NewRequest(http.MethodPost, c.projectsBase, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.do(req)
	if err != nil {
		return "", err
	}

	var project projectResponse
	if err := json.Unmarshal(body, &project); err != nil {
		return "", fmt.Errorf("could not decode project response: %w", err)
	}
	return project.ID, nil
}

func (c *modtaleClient) updateProjectMetadata(projectID string, input projectMetadataInput) error {
	payload := map[string]any{}
	addString(payload, "title", input.title)
	addString(payload, "slug", input.slug)
	addString(payload, "description", input.description)
	addString(payload, "about", input.about)
	addString(payload, "repositoryUrl", input.repositoryURL)
	addString(payload, "license", input.license)
	addString(payload, "hmWikiSlug", input.hmWikiSlug)
	if len(input.tags) > 0 {
		payload["tags"] = []string(input.tags)
	}
	if len(input.links) > 0 {
		payload["links"] = map[string]string(input.links)
	}
	addOptionalBool(payload, "allowModpacks", input.allowModpacks)
	addOptionalBool(payload, "allowComments", input.allowComments)
	addOptionalBool(payload, "hmWikiEnabled", input.hmWikiEnabled)
	addOptionalBool(payload, "galleryCarouselEnabled", input.galleryCarouselEnabled)

	if len(payload) == 0 {
		fmt.Println("No metadata fields were provided; skipping metadata sync.")
		return nil
	}

	return c.sendJSON(http.MethodPut, joinURL(c.projectsBase, url.PathEscape(projectID)), payload, nil)
}

func (c *modtaleClient) inspectManifest(projectID string, artifact string) (*manifestInspectionResult, error) {
	body, err := c.sendMultipart(
		http.MethodPost,
		joinURL(c.projectsBase, url.PathEscape(projectID), "versions", "dependency-suggestions"),
		"file",
		artifact,
		nil,
	)
	if err != nil || c.dryRun {
		return nil, err
	}

	var result manifestInspectionResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("could not decode dependency suggestions: %w", err)
	}
	return &result, nil
}

func (c *modtaleClient) uploadVersion(projectID string, artifact string, version string, gameVersions []string, dependencies []string, incompatibleProjects []string, changelog string, channel string, replaceExisting bool) error {
	fields := map[string][]string{
		"versionNumber": []string{version},
	}
	addFieldList(fields, "gameVersions", gameVersions)
	addFieldList(fields, "modIds", dependencies)
	addFieldList(fields, "incompatibleProjectIds", incompatibleProjects)
	addField(fields, "changelog", changelog)
	addField(fields, "channel", strings.ToUpper(strings.TrimSpace(channel)))
	if replaceExisting {
		addField(fields, "replaceExisting", "true")
	}

	_, err := c.sendMultipart(
		http.MethodPost,
		joinURL(c.projectsBase, url.PathEscape(projectID), "versions"),
		"file",
		artifact,
		fields,
	)
	return err
}

func (c *modtaleClient) updateVersion(projectID string, versionID string, gameVersions []string, dependencies []string, incompatibleProjects []string, changelog string, channel string) error {
	payload := map[string]any{}
	if len(gameVersions) > 0 {
		payload["gameVersions"] = []string(gameVersions)
	}
	if len(dependencies) > 0 {
		payload["modIds"] = []string(dependencies)
	}
	if len(incompatibleProjects) > 0 {
		payload["incompatibleProjectIds"] = []string(incompatibleProjects)
	}
	addString(payload, "changelog", changelog)
	addString(payload, "channel", strings.ToUpper(strings.TrimSpace(channel)))
	if len(payload) == 0 {
		fmt.Println("No version metadata fields were provided; skipping version update.")
		return nil
	}

	return c.sendJSON(
		http.MethodPut,
		joinURL(c.projectsBase, url.PathEscape(projectID), "versions", url.PathEscape(versionID)),
		payload,
		nil,
	)
}

func (c *modtaleClient) uploadProjectFile(projectID string, method string, route string, filePath string) error {
	_, err := c.sendMultipart(
		method,
		joinURL(c.projectsBase, url.PathEscape(projectID), route),
		"file",
		filePath,
		nil,
	)
	return err
}

func (c *modtaleClient) addGalleryVideo(projectID string, videoURL string) error {
	return c.sendJSON(
		http.MethodPost,
		joinURL(c.projectsBase, url.PathEscape(projectID), "gallery", "youtube"),
		map[string]string{"videoUrl": videoURL},
		nil,
	)
}

func (c *modtaleClient) projectTransition(projectID string, transition string) error {
	endpoint := joinURL(c.projectsBase, url.PathEscape(projectID), transition)
	if c.dryRun {
		fmt.Printf("DRY RUN POST %s\n", endpoint)
		return nil
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		nil,
	)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

func (c *modtaleClient) sendJSON(method string, endpoint string, payload any, result any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if c.dryRun {
		fmt.Printf("DRY RUN %s %s\n%s\n", method, endpoint, string(encoded))
		return nil
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := c.do(req)
	if err != nil {
		return err
	}
	if result != nil && len(body) > 0 {
		return json.Unmarshal(body, result)
	}
	return nil
}

func (c *modtaleClient) sendMultipart(method string, endpoint string, fileField string, filePath string, fields map[string][]string) ([]byte, error) {
	if c.dryRun {
		fmt.Printf("DRY RUN %s %s\n", method, endpoint)
		if fileField != "" {
			fmt.Printf("  %s=@%s\n", fileField, filePath)
		}
		for key, values := range fields {
			for _, value := range values {
				fmt.Printf("  %s=%s\n", key, value)
			}
		}
		return nil, nil
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if fileField != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile(fileField, filepath.Base(filePath))
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, file); err != nil {
			return nil, err
		}
	}

	for key, values := range fields {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				return nil, err
			}
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, endpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return c.do(req)
}

func (c *modtaleClient) do(req *http.Request) ([]byte, error) {
	if c.apiKey != "" {
		req.Header.Set("X-MODTALE-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(string(body))
		if len(message) > 2000 {
			message = message[:2000] + "..."
		}
		return nil, fmt.Errorf("%s %s failed: %s\n%s", req.Method, req.URL.String(), resp.Status, message)
	}
	return body, nil
}

func addString(payload map[string]any, key string, value string) {
	if strings.TrimSpace(value) != "" {
		payload[key] = value
	}
}

func addOptionalBool(payload map[string]any, key string, value optionalBool) {
	if value.set {
		payload[key] = value.value
	}
}

func addField(fields map[string][]string, key string, value string) {
	if strings.TrimSpace(value) != "" {
		fields[key] = append(fields[key], value)
	}
}

func addFieldList(fields map[string][]string, key string, values []string) {
	for _, value := range values {
		addField(fields, key, value)
	}
}

func metadataHasFields(title string, summary string, slug string, about string, repositoryURL string, license string, hmWikiSlug string, tags []string, links map[string]string, bools ...optionalBool) bool {
	if title != "" || summary != "" || slug != "" || about != "" || repositoryURL != "" || license != "" || hmWikiSlug != "" || len(tags) > 0 || len(links) > 0 {
		return true
	}
	for _, value := range bools {
		if value.set {
			return true
		}
	}
	return false
}

func normalizeAPIEndpoints(baseURL string, legacyEndpoint string) (string, string) {
	apiBase := strings.TrimRight(baseURL, "/")
	projectsBase := joinURL(apiBase, "projects")
	if legacyEndpoint != "" {
		projectsBase = strings.TrimRight(legacyEndpoint, "/")
		apiBase = strings.TrimSuffix(projectsBase, "/projects")
	}
	if strings.HasSuffix(apiBase, "/projects") {
		projectsBase = apiBase
		apiBase = strings.TrimSuffix(apiBase, "/projects")
	}
	return apiBase, projectsBase
}

func joinURL(base string, parts ...string) string {
	out := strings.TrimRight(base, "/")
	for _, part := range parts {
		out += "/" + strings.Trim(part, "/")
	}
	return out
}

func findDefaultArtifact() (string, error) {
	patterns := []string{
		"build/libs/*.jar",
		"build/libs/*.zip",
		"dist/*.zip",
	}

	var candidates []string
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		candidates = append(candidates, matches...)
	}
	sort.Strings(candidates)
	for _, candidate := range candidates {
		if !strings.Contains(filepath.Base(candidate), "-plain.") {
			return candidate, nil
		}
	}
	if len(candidates) > 0 {
		return candidates[0], nil
	}
	return "", fmt.Errorf("no artifact found; pass --file or build into build/libs first")
}

func versionFromArtifact(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func printManifestSuggestions(result manifestInspectionResult) {
	if result.ModVersion != "" {
		fmt.Printf("Manifest version: %s\n", result.ModVersion)
	}
	if result.GameVersion != "" {
		fmt.Printf("Manifest game version: %s\n", result.GameVersion)
	}
	if len(result.Suggestions) == 0 {
		fmt.Println("No dependency suggestions were returned.")
		return
	}
	fmt.Println("Dependency suggestions:")
	for _, suggestion := range result.Suggestions {
		if suggestion.DependencyEntry != "" {
			fmt.Printf("  --dependency %s\n", suggestion.DependencyEntry)
		}
	}
}

func envDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envBool(name string, fallback bool) bool {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func requireAPIKey(client *modtaleClient) {
	if client.apiKey == "" && !client.dryRun {
		log.Fatalf("MODTALE_KEY, MODTALE_API_KEY, or --api-key is required")
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
