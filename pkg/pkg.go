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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	cron "github.com/robfig/cron/v3"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
)

type MopherConfig struct {
	Writer         io.Writer
	Output         string //creates writer if none is set
	Org            string
	MaxConn        int
	Graph          string
	Verbose        bool
	Dep            string
	WarnUnsyncDev  bool
	PreOutputHook  PreOutputHookFunction
	OutputTemplate string
	OutputEncode   string
}

type PreOutputHookFunction = func(warnings string) (changedWarnings string, shouldBeWritenToOutput bool)

func CronMopher(ctx context.Context, cronString string, config MopherConfig) error {
	c := cron.New()
	_, err := c.AddFunc(cronString, func() {
		err := Mopher(config)
		if err != nil {
			log.Println("ERROR:", err)
		}
	})
	if err != nil {
		return err
	}

	//initial run
	err = Mopher(config)
	if err != nil {
		return err
	}

	c.Start()
	go func() {
		<-ctx.Done()
		c.Stop()
	}()
	return nil
}

func Mopher(config MopherConfig) error {
	if config.Org == "" {
		return errors.New("missing org input")
	}

	tmpl, err := template.New("templ").Parse(config.OutputTemplate)
	if err != nil {
		return err
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
		var templateInput string
		switch config.OutputEncode {
		case "application/json":
			temp, err := json.Marshal(warnings)
			if err != nil {
				return err
			}
			templateInput = string(temp)
		case "plain/text":
			templateInput = warnings
		default:
			templateInput = warnings
		}

		templateOutBuff := strings.Builder{}
		err = tmpl.Execute(&templateOutBuff, map[string]interface{}{"Output": templateInput})
		if err != nil {
			return err
		}
		templateOutput := templateOutBuff.String()
		switch {
		case config.Writer != nil:
			_, err = config.Writer.Write([]byte(templateOutput))
		case strings.HasPrefix(config.Output, "http://") || strings.HasPrefix(config.Output, "https://"):
			err = SendSlackNotification(config.Output, templateOutput)
		case config.Output == "":
			fmt.Print(templateOutput)
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
			_, err = file.Write([]byte(templateOutput))
			if err != nil {
				return fmt.Errorf("unable to open write to output file %v %w", config.Output, err)
			}
		}
	}

	return nil
}
