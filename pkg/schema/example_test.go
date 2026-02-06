/**
 * MIT License
 *
 * Copyright (c) 2025 Stefan Nuxoll
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

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
