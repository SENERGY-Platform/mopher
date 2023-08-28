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
	"strings"
)

func (this *Parsed) PrintDependencyVersionWarnings() error {
	for dep, _ := range this.Latest {
		err := this.PrintVersionWarningsForDependency(dep)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Parsed) PrintVersionWarningsForDependency(dep string) error {
	latestVersion := this.Latest[dep]
	list, err := this.listOldDependencyVersionUsage(dep, latestVersion)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		fmt.Printf("\n\nthe following repositories use a %v version != %v %v\n", dep, latestVersion.Hash, latestVersion.LatestTag)
	}
	slices.SortFunc(list, func(a, b VersionUsageRef) int {
		return strings.Compare(a.Version, b.Version)
	})
	for _, e := range list {
		fmt.Println(e.Version, e.Name)
	}
	return nil
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
