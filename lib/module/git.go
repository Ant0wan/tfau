package module

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"                  // Git plumbing types
	"github.com/go-git/go-git/v5/plumbing/transport"        // Git transport protocols
	"github.com/go-git/go-git/v5/plumbing/transport/client" // Git client
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"    // SSH transport
	"github.com/hashicorp/go-version"                       // Semantic version parsing
)

// fetchGitTags fetches all tags from a Git repository without cloning it.
func fetchGitTags(source string) ([]string, error) {
	// Parse the repository URL
	ep, err := transport.NewEndpoint(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository URL: %v", err)
	}

	// Create a Git client
	gitClient, err := client.NewClient(ep)
	if err != nil {
		return nil, fmt.Errorf("failed to create Git client: %v", err)
	}

	// Open a session to the remote repository
	session, err := gitClient.NewUploadPackSession(ep, getGitAuth(source))
	if err != nil {
		return nil, fmt.Errorf("failed to create upload pack session: %v", err)
	}
	defer session.Close()

	// Fetch the advertised references (including tags)
	refs, err := session.AdvertisedReferences()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch advertised references: %v", err)
	}

	// Extract tag names
	var tags []string
	for refName := range refs.References {
		// Convert refName to plumbing.ReferenceName
		ref := plumbing.ReferenceName(refName)
		if ref.IsTag() {
			tags = append(tags, ref.Short())
		}
	}

	return tags, nil
}

// getGitAuth returns the appropriate authentication method based on the Git URL.
func getGitAuth(source string) transport.AuthMethod {
	if strings.HasPrefix(source, "https://") {
		// No authentication for public HTTPS repositories
		return nil
	} else if strings.HasPrefix(source, "ssh://") || strings.HasPrefix(source, "git@") {
		// Use SSH authentication with the SSH agent
		authMethod, err := ssh.DefaultAuthBuilder("git")
		if err != nil {
			log.Fatalf("Failed to create SSH auth method: %v", err)
		}
		return authMethod
	}
	return nil
}

// getLatestVersionFromGit retrieves the latest version from a Git repository using the Go Git library.
func getLatestVersionFromGit(source string) (string, error) {
	// Fetch all tags from the Git repository
	tags, err := fetchGitTags(source)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Git tags: %v", err)
	}

	// Parse tags into semantic version objects
	versions := make([]*version.Version, 0, len(tags))
	for _, tag := range tags {
		parsedVersion, err := version.NewVersion(tag)
		if err != nil {
			log.Printf("Warning: Skipping invalid version %s: %v", tag, err)
			continue
		}
		versions = append(versions, parsedVersion)
	}

	// Sort versions in descending order
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})

	// Log all versions
	versionStrings := make([]string, 0, len(versions))
	for _, v := range versions {
		versionStrings = append(versionStrings, v.String())
	}
	log.Printf("All versions of module %s: %v", source, versionStrings)

	// Log the latest version
	if len(versions) == 0 {
		return "", fmt.Errorf("no valid versions found for module: %s", source)
	}
	latestVersion := versions[0].String()
	log.Printf("Latest version of module %s: %s", source, latestVersion)

	return latestVersion, nil
}
