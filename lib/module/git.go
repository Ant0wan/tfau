package module

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"                    // Updated import
	"github.com/go-git/go-git/v5/plumbing"           // Updated import
	"github.com/go-git/go-git/v5/plumbing/transport" // Updated import

	// Updated import
	"github.com/go-git/go-git/v5/plumbing/transport/ssh" // Updated import
	"github.com/go-git/go-git/v5/storage/memory"         // Updated import
	"github.com/hashicorp/go-version"                    // Semantic version parsing
)

// fetchGitTags fetches all tags from a Git repository using the Go Git library.
func fetchGitTags(source string) ([]string, error) {
	// Create a new in-memory repository
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           source,
		ReferenceName: "refs/tags/*",
		SingleBranch:  false,
		Depth:         1,
		Auth:          getGitAuth(source), // Get authentication based on the protocol
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %v", err)
	}

	// Fetch all tags
	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %v", err)
	}

	// Extract tag names
	var tags []string
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		tags = append(tags, tagName)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate over tags: %v", err)
	}

	return tags, nil
}

// getGitAuth returns the appropriate authentication method based on the Git URL.
func getGitAuth(source string) transport.AuthMethod {
	if strings.HasPrefix(source, "https://") {
		// No authentication for public HTTPS repositories
		return nil
	} else if strings.HasPrefix(source, "ssh://") || strings.HasPrefix(source, "git@") {
		// Use SSH authentication
		publicKeys, err := ssh.NewPublicKeysFromFile("git", os.ExpandEnv("$HOME/.ssh/id_rsa"), "")
		if err != nil {
			log.Fatalf("Failed to load SSH key: %v", err)
		}
		return publicKeys
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
