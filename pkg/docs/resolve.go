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

package docs

import (
	"fmt"
	"strings"

	"go-valkyrie.com/odin/pkg/model"
)

// ResolveReference matches a reference string against a list of component templates.
// Resolution order:
//  1. Fully qualified (contains ":#") — exact match
//  2. Definition name — strip "#", case-insensitive; unique match wins
//  3. Package name — last path segment, strip @vN, case-insensitive; unique def in package wins
//  4. Dot syntax "package.Definition" — case-insensitive match on both parts
//  5. Error with available templates
func ResolveReference(reference string, templates []*model.ComponentTemplate) (*model.ComponentTemplate, error) {
	if len(templates) == 0 {
		return nil, fmt.Errorf("no component templates available")
	}

	// 1. Fully qualified match: package:#Definition
	if strings.Contains(reference, ":#") {
		parts := strings.SplitN(reference, ":#", 2)
		pkg, defName := parts[0], "#"+parts[1]
		for _, tmpl := range templates {
			if tmpl.Package == pkg && tmpl.Name == defName {
				return tmpl, nil
			}
		}
		for _, tmpl := range templates {
			fullPkg := tmpl.Package + "@" + tmpl.Version
			if fullPkg == pkg && tmpl.Name == defName {
				return tmpl, nil
			}
		}
		return nil, fmt.Errorf("no component template found matching %q", reference)
	}

	refLower := strings.ToLower(reference)

	// 2. Definition name match (strip "#", case-insensitive)
	var defMatches []*model.ComponentTemplate
	for _, tmpl := range templates {
		defName := strings.TrimPrefix(tmpl.Name, "#")
		if strings.EqualFold(defName, refLower) {
			defMatches = append(defMatches, tmpl)
		}
	}
	if len(defMatches) == 1 {
		return defMatches[0], nil
	}

	// 3. Package name match (last path segment, strip @vN, case-insensitive)
	var pkgMatches []*model.ComponentTemplate
	for _, tmpl := range templates {
		if strings.EqualFold(shorthandName(tmpl.Package), refLower) {
			pkgMatches = append(pkgMatches, tmpl)
		}
	}
	if len(pkgMatches) == 1 {
		return pkgMatches[0], nil
	}

	// 4. Dot syntax: "package.Definition" (case-insensitive)
	if dotIdx := strings.LastIndex(reference, "."); dotIdx != -1 {
		pkgPart := strings.ToLower(reference[:dotIdx])
		defPart := strings.ToLower(reference[dotIdx+1:])
		for _, tmpl := range templates {
			pkg := strings.ToLower(shorthandName(tmpl.Package))
			def := strings.ToLower(strings.TrimPrefix(tmpl.Name, "#"))
			if pkg == pkgPart && def == defPart {
				return tmpl, nil
			}
		}
	}

	// 5. Error with available templates
	// If we had ambiguous matches from step 2 (definition name), show those as candidates
	if len(defMatches) > 1 {
		return nil, ambiguousError(reference, defMatches)
	}
	// If we had ambiguous matches from step 3 (package name), show those as candidates
	if len(pkgMatches) > 1 {
		return nil, ambiguousError(reference, pkgMatches)
	}

	// No match at all
	available := make([]string, 0, len(templates))
	for _, tmpl := range templates {
		available = append(available, displayName(tmpl))
	}
	return nil, fmt.Errorf("no component template matching %q; available: %s", reference, strings.Join(available, ", "))
}

func ambiguousError(reference string, matches []*model.ComponentTemplate) error {
	candidates := make([]string, 0, len(matches))
	for _, m := range matches {
		candidates = append(candidates, fmt.Sprintf("%s (%s)", displayName(m), m.Package))
	}
	return fmt.Errorf("ambiguous reference %q matches multiple templates:\n  %s", reference, strings.Join(candidates, "\n  "))
}

func displayName(tmpl *model.ComponentTemplate) string {
	return shorthandName(tmpl.Package) + "." + strings.TrimPrefix(tmpl.Name, "#")
}

// shorthandName extracts the last path segment from a package import path,
// stripping any @vN version suffix.
func shorthandName(pkg string) string {
	// Strip @vN suffix if present
	if idx := strings.LastIndex(pkg, "@"); idx != -1 {
		pkg = pkg[:idx]
	}
	// Get last path segment
	if idx := strings.LastIndex(pkg, "/"); idx != -1 {
		return pkg[idx+1:]
	}
	return pkg
}
