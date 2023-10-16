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

type MopherConfig struct {
	Writer        io.Writer
	Output        string //creates writer if none is set
	Org           string
	MaxConn       int
	Graph         string
	Verbose       bool
	Dep           string
	WarnUnsyncDev bool
	PreOutputHook PreOutputHookFunction
}

type PreOutputHookFunction = func(warnings string) (changedWarnings string, shouldBeWritenToOutput bool)

func Mopher(config MopherConfig) error {
	if config.Org == "" {
		return errors.New("missing org input")
	}

	parsed, err := LoadOrg(config.Org, config.MaxConn)
	if err != nil {
		return err
	}

	var writer strings.Builder
	parsed.SetOutput(&writer)
	defer writer.Reset()

	if config.Graph != "" {
		err = parsed.StoreGraph(config.Graph, config.Verbose)
		if err != nil {
			return err
		}
	}
	if config.Dep != "" {
		err = parsed.PrintDependents(config.Dep)
		if err != nil {
			return err
		}
	}
	err = parsed.PrintWarnings(config.WarnUnsyncDev)
	if err != nil {
		return err
	}

	warnings := writer.String()

	write := true
	if config.PreOutputHook != nil {
		warnings, write = config.PreOutputHook(warnings)
	}

	if write {
		switch {
		case config.Writer != nil:
			_, err = config.Writer.Write([]byte(warnings))
		case strings.HasPrefix(config.Output, "http://") || strings.HasPrefix(config.Output, "https://"):
			err = SendSlackNotification(config.Output, warnings)
		case config.Output == "":
			fmt.Print(warnings)
		default:
			var file io.WriteCloser
			file, err = os.OpenFile(config.Output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("unable to open output file %v %w", config.Output, err)
			}
			defer func() {
				err := file.Close()
				if err != nil {
					fmt.Println("unable to close output file", config.Output, err)
				}
			}()
			_, err = config.Writer.Write([]byte(warnings))
			if err != nil {
				return fmt.Errorf("unable to open write to output file %v %w", config.Output, err)
			}
		}
	}

	return nil
}
