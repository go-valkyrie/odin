// SPDX-License-Identifier: MIT

package schema_test

import (
	"context"
	"fmt"
	"os"

	"go-valkyrie.com/odin/pkg/model"
	"go-valkyrie.com/odin/pkg/schema"
)

// ExampleBundle_ValuesSchema demonstrates how to get schema information for a bundle's values.
func ExampleBundle_ValuesSchema() {
	// Load a bundle
	b, err := model.LoadBundle("path/to/bundle")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	// Get the values schema
	fields := b.ValuesSchema()

	// Print the schema to stdout
	schema.FormatSchema(os.Stdout, fields, 0)
}

// ExampleComponentTemplate_ConfigSchema demonstrates how to get schema information for a component template's config.
func ExampleComponentTemplate_ConfigSchema() {
	// Load a bundle
	b, err := model.LoadBundle("path/to/bundle")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	// Iterate through component templates
	ctx := context.Background()
	for tmpl, err := range b.ComponentTemplates(ctx) {
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}

		// Get the config schema with expansion enabled
		fields := tmpl.ConfigSchema(schema.WithExpand(true))

		// Print the schema
		fmt.Printf("# %s %s\n\n", tmpl.Package, tmpl.Name)
		schema.FormatSchema(os.Stdout, fields, 2)
		break // Just show first template for example
	}
}
