/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func Mopher(output string, org string, maxConn int, graph string, verbose bool, dep string, warnUnsyncDev bool) (err error) {
	var writer io.Writer
	switch {
	case strings.HasPrefix(output, "http://") || strings.HasPrefix(output, "https://"):
		temp := NewSlackWriter(output, verbose)
		defer func() {
			if err == nil {
				err = temp.Close()
				if err != nil {
					err = fmt.Errorf("unable to send output %w", err)
				}
			}
		}()
		writer = io.MultiWriter(temp, os.Stdout)
	case output == "":
		writer = os.Stdout
	default:
		var file io.WriteCloser
		file, err = os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("unable to open output file %v %w", output, err)
		}
		writer = file
		defer func() {
			if err == nil {
				err = file.Close()
				if err != nil {
					err = fmt.Errorf("unable to close output file %v %w", output, err)
				}
			}
		}()
	}
	return MopherWithWriter(writer, org, maxConn, graph, verbose, dep, warnUnsyncDev)
}

func MopherWithWriter(writer io.Writer, org string, maxConn int, graph string, verbose bool, dep string, warnUnsyncDev bool) error {
	if org == "" {
		return errors.New("missing org input")
	}

	parsed, err := LoadOrg(org, maxConn)
	if err != nil {
		return err
	}

	parsed.SetOutput(writer)

	if graph != "" {
		err = parsed.StoreGraph(graph, verbose)
		if err != nil {
			return err
		}
	}
	if dep != "" {
		err = parsed.PrintDependents(dep)
		if err != nil {
			return err
		}
	}
	err = parsed.PrintWarnings(warnUnsyncDev)
	if err != nil {
		return err
	}
	return nil
}
