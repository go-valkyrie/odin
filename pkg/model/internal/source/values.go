// SPDX-License-Identifier: MIT

package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"go-valkyrie.com/odin/internal/utils"
	"go-valkyrie.com/odin/internal/utils/regexpext"
)

// cueFormat is the location-format tag that marks a values location as an
// independent, self-rooted CUE module rather than part of the bundle's own
// module tree.
const cueFormat = "cue"

var _valuesFilePattern = utils.Must(regexpext.NewMatcher(`^((?P<Format>[\w]*): )?(?P<Path>.*$)`))

type valuesFile struct {
	format string
	path   string
}

func (f *valuesFile) String() string {
	if f.format == "" {
		return f.path
	}
	return fmt.Sprintf("%s: %s", f.format, f.path)
}

// Values is a source for values overlays loaded from one or more files.
type Values struct {
	locations []valuesFile
}

func NewValues(locations []string) (*Values, error) {
	files := make([]valuesFile, 0, len(locations))
	for _, location := range locations {
		if match := _valuesFilePattern.Match(location); match != nil {
			file := valuesFile{
				format: match.Named("Format"),
				path:   match.Named("Path"),
			}
			info, err := os.Stat(file.path)
			if err != nil {
				return nil, err
			}
			if file.format == cueFormat && !info.IsDir() {
				return nil, fmt.Errorf("values location %q: %q locations must be a directory (a CUE module is a directory containing cue.mod)", file.path, file.format)
			}
			files = append(files, file)
		}
	}
	return &Values{locations: files}, nil
}

func (s *Values) String() string {
	sb := strings.Builder{}
	count := len(s.locations)
	for i, file := range s.locations {
		sb.WriteString(file.String())
		if i < count-1 {
			sb.WriteString("; ")
		}
	}
	return sb.String()
}

func (s *Values) Load(ctx *cue.Context, opts *LoadOptions) (cue.Value, error) {
	var defaultFiles, cueFiles []valuesFile
	for _, file := range s.locations {
		if file.format == cueFormat {
			cueFiles = append(cueFiles, file)
		} else {
			defaultFiles = append(defaultFiles, file)
		}
	}

	var result cue.Value
	haveValue := false
	unify := func(v cue.Value) {
		if !haveValue {
			result, haveValue = v, true
			return
		}
		result = result.Unify(v)
	}

	// Default group: locations that are part of the bundle's own module
	// (untagged, or explicitly yaml:/json:). Batched into a single
	// load.Instances call, as before, but ModuleRoot is now pinned to the
	// bundle's own directory so CUE's nested-module check anchors correctly
	// regardless of the process's working directory. Dir is left unset so
	// relative CLI paths keep resolving against the user's shell cwd.
	//
	// Only skipped when there's at least one cue: location AND no default
	// locations, so a cue:-only Values doesn't spuriously also load "." (the
	// process cwd) as an extra default-group package.
	if len(defaultFiles) > 0 || len(cueFiles) == 0 {
		args := make([]string, 0, len(defaultFiles)*2)
		for _, file := range defaultFiles {
			if file.format != "" {
				args = append(args, fmt.Sprintf("%s:", file.format), file.path)
			} else {
				args = append(args, file.path)
			}
		}

		inst := load.Instances(args, &load.Config{
			DataFiles:  true,
			Env:        opts.Env,
			ModuleRoot: opts.ModuleRoot,
		})[0]

		if configure := opts.InstanceConfiguration; configure != nil {
			if err := configure(inst); err != nil {
				return cue.Value{}, err
			}
		}

		unify(ctx.BuildInstance(inst))
	}

	// cue: group: each location is its own independent, self-rooted CUE
	// module, decoupled from the bundle and from each other, but still
	// sharing the bundle's registry configuration via opts.Env.
	//
	// A cue: module's top-level fields ARE the values content directly
	// (e.g. "components: api: image: ..."), like yaml:/json: locations --
	// not wrapped in a "values: {...}" field the way a .cue overlay file
	// living inside the bundle's own module can be. That wrapper only
	// works for same-module files because CUE resolves them as sharing the
	// bundle's own package identity; FillPath-ing a value from a genuinely
	// separate module doesn't deep-merge across a "values" key the same
	// way, it just adds an inert sibling field.
	for _, file := range cueFiles {
		absPath, err := filepath.Abs(file.path)
		if err != nil {
			return cue.Value{}, err
		}

		inst := load.Instances([]string{"."}, &load.Config{
			Dir:        absPath,
			ModuleRoot: absPath,
			DataFiles:  true,
			Env:        opts.Env,
		})[0]

		if configure := opts.InstanceConfiguration; configure != nil {
			if err := configure(inst); err != nil {
				return cue.Value{}, err
			}
		}

		unify(ctx.BuildInstance(inst))
	}

	return result, nil
}
