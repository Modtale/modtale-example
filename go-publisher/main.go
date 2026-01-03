package main

import (
    "bytes"
    "flag"
    "fmt"
    "io"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
)

func main() {
    // CLI flags
    jarPath := flag.String("jar", "", "Path to the JAR file to upload (default tries build/libs/*.jar)")
    version := flag.String("version", "", "versionNumber field to send (default: jar base name)")
    gameVersions := flag.String("gameVersions", "Release 1.0", "gameVersions form field")
    changelog := flag.String("changelog", "Uploaded via Go publisher", "changelog form field")
    endpoint := flag.String("endpoint", "https://api.modtale.net/api/v1/projects", "base API endpoint")
    flag.Parse()

    apiKey := os.Getenv("MODTALE_KEY")
    projectID := os.Getenv("MODTALE_PROJECT_ID")
    if apiKey == "" || projectID == "" {
       log.Fatalf("MODTALE_KEY and MODTALE_PROJECT_ID environment variables are required")
    }

    // Determine jar
    jar := *jarPath
    if jar == "" {
       // try to find a jar in build/libs
       candidates, err := filepath.Glob("build/libs/*.jar")
       if err != nil || len(candidates) == 0 {
          log.Fatalf("No jar found in build/libs and --jar not provided. Build the project first or pass --jar.")
       }
       jar = candidates[0]
    }

    jarFile, err := os.Open(jar)
    if err != nil {
       log.Fatalf("Failed to open jar file %s: %v", jar, err)
    }
    defer jarFile.Close()

    // default version: jar base name if not provided
    actualVersion := *version
    if actualVersion == "" {
       actualVersion = filepath.Base(jar)
    }

    // Prepare multipart form
    var body bytes.Buffer
    writer := multipart.NewWriter(&body)

    // file field
    part, err := writer.CreateFormFile("file", filepath.Base(jar))
    if err != nil {
       log.Fatalf("CreateFormFile error: %v", err)
    }
    _, err = io.Copy(part, jarFile)
    if err != nil {
       log.Fatalf("Copy jar to form error: %v", err)
    }

    // other fields
    _ = writer.WriteField("versionNumber", actualVersion)
    _ = writer.WriteField("gameVersions", *gameVersions)
    _ = writer.WriteField("changelog", *changelog)

    if err := writer.Close(); err != nil {
       log.Fatalf("closing multipart writer: %v", err)
    }

    url := fmt.Sprintf("%s/%s/versions", *endpoint, projectID)
    req, err := http.NewRequest("POST", url, &body)
    if err != nil {
       log.Fatalf("NewRequest error: %v", err)
    }

    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("X-MODTALE-KEY", apiKey)

    // do request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
       log.Fatalf("HTTP request failed: %v", err)
    }
    defer resp.Body.Close()

    respBody, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
       fmt.Printf("Upload successful! Status: %s\nResponse:\n%s\n", resp.Status, string(respBody))
    } else {
       log.Fatalf("Upload failed. Status: %s\nResponse:\n%s\n", resp.Status, string(respBody))
    }

}
