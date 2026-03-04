// SPDX-License-Identifier: MIT

package source

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"go-valkyrie.com/odin/internal/utils"
	"go-valkyrie.com/odin/internal/utils/regexpext"
)

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
			if _, err := os.Stat(file.path); err != nil {
				return nil, err
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
	args := make([]string, 0, len(s.locations)*2)
	for _, file := range s.locations {
		if file.format != "" {
			args = append(args, fmt.Sprintf("%s:", file.format), file.path)
		} else {
			args = append(args, file.path)
		}
	}

	inst := load.Instances(args, &load.Config{
		DataFiles: true,
		Env:       opts.Env,
	})[0]

	if configure := opts.InstanceConfiguration; configure != nil {
		if err := configure(inst); err != nil {
			return cue.Value{}, err
		}
	}

	return ctx.BuildInstance(inst), nil
}
