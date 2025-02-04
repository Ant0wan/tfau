package hcl

import (
	"fmt"
		"github.com/zclconf/go-cty/cty"
	"io/ioutil"

	"tfau/lib/module"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)


func UpdateVersionInFile(filename string) error {
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
