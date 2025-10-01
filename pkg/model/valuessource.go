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

package model

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
	} else {
		return fmt.Sprintf("%s: %s", f.format, f.path)
	}
}

type valuesSource struct {
	locations []valuesFile
}

func newValuesSource(locations []string) (*valuesSource, error) {
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

	return &valuesSource{locations: files}, nil
}

func (s *valuesSource) String() string {
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

func (s *valuesSource) Load(ctx *cue.Context, opts *sourceLoadOptions) (cue.Value, error) {
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
