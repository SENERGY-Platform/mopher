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
	org     string
}

type InverseIndexModRef struct {
	UsesVersion     string
	UserModule      string
	SemanticVersion bool
}

type LatestCommitInfo struct {
	Hash      string
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
	list := this.Inverse[dep]
	if len(list) > 0 {
		fmt.Printf("\n\n%v is used by the following repositories (sorted by usage-version)\n", dep)
		slices.SortFunc(list, func(a, b InverseIndexModRef) int {
			return strings.Compare(a.UserModule, b.UserModule)
		})
		slices.SortFunc(list, func(a, b InverseIndexModRef) int {
			return strings.Compare(a.UsesVersion, b.UsesVersion)
		})
		for _, ref := range list {
			fmt.Println(ref.UserModule, ref.UsesVersion)
		}
	} else {
		fmt.Printf("\n\n%v is used by no %v repository as dependency\n", dep, this.org)
	}

	return nil
}

func (this *Parsed) PrintWarnings() error {
	updateOrderFilter := map[string]bool{}
	deprecated, err := this.PrintWrongModuleNameWarnings()
	if err != nil {
		return err
	}
	for _, d := range deprecated {
		updateOrderFilter[d] = true
	}
	deprecated, err = this.PrintGoVersionWarnings()
	if err != nil {
		return err
	}
	for _, d := range deprecated {
		updateOrderFilter[d] = true
	}
	deprecated, err = this.PrintDependencyVersionWarnings()
	if err != nil {
		return err
	}
	for _, d := range deprecated {
		updateOrderFilter[d] = true
	}
	err = this.PrintUpdateOrder(updateOrderFilter)
	return nil
}

func (this *Parsed) PrintWrongModuleNameWarnings() (deprecated []string, err error) {
	invalidNames := []string{}
	for name, _ := range this.Modules {
		if !strings.HasPrefix(name, GithubUrl+"/"+this.org) {
			invalidNames = append(invalidNames, name)
		}
	}
	if len(invalidNames) > 0 {
		fmt.Println("\n\nfound unexpected module names:")
		for _, name := range invalidNames {
			fmt.Println(name)
			deprecated = append(deprecated, name)
		}
	}
	return deprecated, nil
}

func (this *Parsed) PrintUpdateOrder(filter map[string]bool) error {
	order, err := this.GetRecommendedUpdateOrder()
	if err != nil {
		return err
	}
	fmt.Printf("\n\nrecommended update order:\n")
	for _, e := range order {
		if this.toBeUpdated(filter, e) {
			filter[e] = true
			fmt.Println(e)
		}
	}
	return nil
}

type VersionUsageRef struct {
	Name    string
	Version string
}
