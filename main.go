package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type ModuleInfo struct {
	Versions []struct {
		Version string `json:"version"`
	} `json:"modules"`
}

func getLatestVersion(source string) (string, error) {
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

func updateVersionInFile(filename string) error {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return fmt.Errorf("failed to parse HCL file: %v", diags)
	}

	syntaxFile, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return fmt.Errorf("failed to cast HCL body to syntax body")
	}

	for _, block := range syntaxFile.Blocks {
		if block.Type == "module" {
			var sourceValue string
			var versionAttr *hclsyntax.Attribute

			for name, attr := range block.Body.Attributes {
				if name == "source" {
					sourceVal, _ := attr.Expr.Value(nil)
					sourceValue = sourceVal.AsString()
				}
				if name == "version" {
					versionAttr = attr
				}
			}

			if sourceValue != "" && versionAttr != nil {
				newVersion, err := getLatestVersion(sourceValue)
				if err == nil {
					versionAttr.Expr = &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal(newVersion),
					}
				}
			}
		}
	}

	output := hclwrite.Format([]byte(filename))
	return ioutil.WriteFile(filename, output, 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <terraform_file.tf>")
		return
	}

	filename := os.Args[1]
	if err := updateVersionInFile(filename); err != nil {
		fmt.Println("Error:", err)
	}
}
