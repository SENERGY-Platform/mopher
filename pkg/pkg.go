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
	"fmt"
	"github.com/google/go-github/v54/github"
	"golang.org/x/mod/modfile"
	"os"
	"slices"
	"strings"
)

type Parsed struct {
	Repos   []*github.Repository
	Modules map[string]*modfile.File
	Inverse map[string][]InverseIndexModRef
	Latest  map[string]LatestCommitInfo
}

type InverseIndexModRef struct {
	UsesVersion     string
	UserModule      string
	SemanticVersion bool
}

type LatestCommitInfo struct {
	Hash      string
	Branch    string
	LatestTag string
}

func (this *Parsed) StoreGraph(outputFile string, verbose bool) error {
	text, err := this.generatePlantuml(verbose)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(text))
	if err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func (this *Parsed) PrintDependents(dep string) error {
	fmt.Printf("\n\n%v is used by the following repositories (sorted by usage-version)\n", dep)
	list := this.Inverse[dep]
	slices.SortFunc(list, func(a, b InverseIndexModRef) int {
		return strings.Compare(a.UserModule, b.UserModule)
	})
	slices.SortFunc(list, func(a, b InverseIndexModRef) int {
		return strings.Compare(a.UsesVersion, b.UsesVersion)
	})
	for _, ref := range list {
		fmt.Println(ref.UserModule, ref.UsesVersion)
	}
	return nil
}

func (this *Parsed) PrintWarnings() error {
	err := this.PrintGoVersionWarnings()
	if err != nil {
		return err
	}
	err = this.PrintDependencyVersionWarnings()
	if err != nil {
		return err
	}
	return nil
}

type VersionUsageRef struct {
	Name    string
	Version string
}
