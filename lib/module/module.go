package module

import (
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strings"
	"fmt"

)

type ModuleInfo struct {
	Versions []struct {
		Version string `json:"version"`
	} `json:"modules"`
}

func GetLatestVersion(source string) (string, error) {
	parts := strings.Split(source, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid module source format")
	}

	registryURL := fmt.Sprintf("https://registry.terraform.io/v1/modules/%s/%s/%s/versions", parts[0], parts[1], parts[2])
	resp, err := http.Get(registryURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch version data")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var moduleInfo ModuleInfo
	if err := json.Unmarshal(body, &moduleInfo); err != nil {
		return "", err
	}

	if len(moduleInfo.Versions) == 0 {
		return "", fmt.Errorf("no versions found")
	}

	latestVersion := moduleInfo.Versions[0].Version
	versionParts := strings.Split(latestVersion, ".")
	if len(versionParts) < 2 {
		return "", fmt.Errorf("invalid version format")
	}

	return fmt.Sprintf("~>%s.%s", versionParts[0], versionParts[1]), nil
}
