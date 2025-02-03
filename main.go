package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"tfau/lib"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)


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
				newVersion, err := module.GetLatestVersion(sourceValue)
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
