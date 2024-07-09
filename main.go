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

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/SENERGY-Platform/mopher/pkg"
	"golang.org/x/mod/modfile"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"syscall"
)

func main() {
	var umod, umodeExecute, umodeInternal, umodeInternalExecute bool
	var org, dep, graph, output, outputTemplate, outputEncode, cron string
	var verbose, warnUnsyncDev, warnGoVersion, distinct bool
	var maxConn int

	flag.BoolVar(&umod, "u", false, "update mode: check local repository for updates and print go get commands")
	flag.BoolVar(&umodeInternal, "ui", false, "update mode: check local repository for updates and print go get commands (without go get -u)")
	flag.BoolVar(&umodeExecute, "ux", false, "update mode: check local repository for updates and execute go get commands")
	flag.BoolVar(&umodeInternalExecute, "uix", false, "update mode: check local repository for updates and execute go get commands (without go get -u)")

	flag.StringVar(&org, "org", "", "github org to be scanned")
	flag.StringVar(&output, "output", "", "output, defaults to std-out; may be a file location or a url")
	flag.StringVar(&outputTemplate, "output_template", "{{.Output}}", "template for output")
	flag.StringVar(&outputEncode, "output_encode", "plain/text", "encode output as plain/text or application/json")
	flag.StringVar(&dep, "dep", "", "dependency to be scanned for in org (optional)")
	flag.StringVar(&graph, "graph", "", "output file for plantuml dependency graph (optional)")
	flag.BoolVar(&verbose, "graph_verbose", false, "include none org dependencies in plantuml")
	flag.StringVar(&cron, "cron", "", "run repeatedly")
	flag.BoolVar(&distinct, "distinct", false, "only output if output has changed (useful for cron jobs)")
	flag.BoolVar(&warnUnsyncDev, "warn_unsync_dev", true, "warn if dev and master/main branches are not at the same commit")
	flag.BoolVar(&warnGoVersion, "warn_go_version", false, "warn if used go version is not the newest version")
	flag.IntVar(&maxConn, "max_conn", 25, "max parallel connections to github")

	flag.BoolFunc("debug", "enables debug logs", func(s string) error {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
		return nil
	})
	flag.Parse()

	//set args by environment variable, if the environment variable is not empty
	flag.VisitAll(func(f *flag.Flag) {
		env := os.Getenv(argNameToEnvName(f.Name))
		if env != "" {
			fmt.Printf("set arg %v by env %v\n", f.Name, env)
			err := f.Value.Set(env)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	if umod || umodeExecute || umodeInternal || umodeInternalExecute {
		runUpdateMode(umodeExecute || umodeInternalExecute, umodeInternal || umodeInternalExecute)
		return
	}

	args := flag.Args()
	switch len(args) {
	case 0:
		if org == "" {
			tempOrg, tempDep, err := getParamsFromDir(".")
			if err != nil {
				log.Fatal(err)
				return
			}
			if org == "" {
				org = tempOrg
			}
			if dep == "" {
				dep = tempDep
			}
		}
	case 1:
		tempOrg, tempDep, err := getParamsFromArg(args[0])
		if err != nil {
			log.Fatal(err)
			return
		}
		if org == "" {
			org = tempOrg
		}
		if dep == "" {
			dep = tempDep
		}
	default:
		log.Fatal("unexpected args", args)
		return
	}

	config := pkg.MopherConfig{
		Output:         output,
		OutputTemplate: outputTemplate,
		OutputEncode:   outputEncode,
		Org:            org,
		MaxConn:        maxConn,
		Graph:          graph,
		Verbose:        verbose,
		Dep:            dep,
		WarnUnsyncDev:  warnUnsyncDev,
		WarnGoVersion:  warnGoVersion,
	}

	if distinct {
		config.PreOutputHook = pkg.GetDistinctHook()
	}

	slog.Debug("Startup", "config", config)

	if cron != "" {
		ctx, cancel := context.WithCancel(context.Background())

		err := pkg.CronMopher(ctx, cron, config)
		if err != nil {
			log.Fatal(err)
		}

		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		sig := <-shutdown
		log.Println("received shutdown signal", sig)
		cancel()
	} else {
		err := pkg.Mopher(config)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func runUpdateMode(execute bool, internal bool) {
	file, err := os.ReadFile("go.mod")
	if err != nil {
		log.Fatal(err)
		return
	}
	mod, err := modfile.ParseLax("go.mod", file, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	commands, err := pkg.RunUpdateMode(mod, internal)
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, cmd := range commands {
		fmt.Printf("%v %v\n", cmd.Cmd, strings.Join(cmd.Args, " "))
		if execute {
			_, err := exec.Command(cmd.Cmd, cmd.Args...).Output()
			if err != nil {
				log.Fatal(err)
				return
			}
		}
	}
}

func getParamsFromArg(arg string) (org string, dep string, err error) {
	arg = strings.TrimPrefix(arg, "http://")
	arg = strings.TrimPrefix(arg, "https://")
	if strings.HasPrefix(arg, pkg.GithubUrl) {
		return getParamsFromGithubUrl(arg)
	} else {
		return getParamsFromDir(arg)
	}
}

func getParamsFromGithubUrl(arg string) (org string, dep string, err error) {
	parts := strings.Split(strings.TrimPrefix(arg, pkg.GithubUrl+"/"), "/")
	switch len(parts) {
	case 0:
		err = errors.New("missing org in github url")
		return
	case 1:
		org = parts[0]
		return
	case 2:
		org, dep = parts[0], arg
		return
	default:
		err = errors.New("unable to parse github url (path is to long)")
		return
	}
}

func getParamsFromDir(arg string) (org string, dep string, err error) {
	file, err := os.ReadFile(path.Join(arg, "go.mod"))
	if err != nil {
		return org, dep, err
	}
	mod, err := modfile.ParseLax("go.mod", file, nil)
	if err != nil {
		return org, dep, err
	}
	return getParamsFromGithubUrl(mod.Module.Mod.Path)
}

var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func argNameToEnvName(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return "MOPHER_" + strings.ToUpper(strings.Join(a, "_"))
}
