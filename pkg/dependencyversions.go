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
	"slices"
	"sort"
	"strings"
)

func (this *Parsed) PrintDependencyVersionWarnings() error {
	//make result deterministic by sorting the keys
	keys := []string{}
	for key, _ := range this.Latest {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	updateOrderFilter := map[string]bool{}

	for _, key := range keys {
		dependent, err := this.PrintVersionWarningsForDependency(key)
		if err != nil {
			return err
		}
		for _, d := range dependent {
			updateOrderFilter[d] = true
		}
	}

	order, err := this.GetRecommendedUpdateOrder()
	if err != nil {
		return err
	}
	fmt.Printf("\n\nrecommended update order:\n")
	for _, e := range order {
		if this.toBeUpdated(updateOrderFilter, e) {
			updateOrderFilter[e] = true
			fmt.Println(e)
		}
	}

	return nil
}

func (this *Parsed) toBeUpdated(filter map[string]bool, e string) bool {
	if filter[e] {
		return true
	}
	module, ok := this.Modules[e]
	if ok {
		for _, req := range module.Require {
			if this.toBeUpdated(filter, req.Mod.Path) {
				return true
			}
		}
	}
	return false
}

func (this *Parsed) PrintVersionWarningsForDependency(dep string) (dependent []string, err error) {
	latestVersion := this.Latest[dep]
	list, err := this.listOldDependencyVersionUsage(dep, latestVersion)
	if err != nil {
		return dependent, err
	}
	if len(list) == 0 {
		return dependent, nil
	}
	fmt.Printf("\n\nthe following repositories use a %v version != %v %v\n", dep, latestVersion.Hash, latestVersion.LatestTag)
	slices.SortFunc(list, func(a, b VersionUsageRef) int {
		result := strings.Compare(a.Version, b.Version)
		if result == 0 {
			result = strings.Compare(a.Name, b.Name)
		}
		return result
	})
	for _, e := range list {
		fmt.Println(e.Version, e.Name)
		dependent = append(dependent, e.Name)
	}
	return dependent, nil
}

func (this *Parsed) listOldDependencyVersionUsage(dep string, version LatestCommitInfo) (result []VersionUsageRef, err error) {
	for _, ref := range this.Inverse[dep] {
		versionStr := version.Hash
		if ref.SemanticVersion {
			versionStr = version.LatestTag
		}
		if ref.UsesVersion != versionStr {
			result = append(result, VersionUsageRef{
				Name:    ref.UserModule,
				Version: ref.UsesVersion,
			})
		}
	}
	return result, nil
}
