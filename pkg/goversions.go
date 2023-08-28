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
	"runtime"
	"slices"
	"strings"
)

func (this *Parsed) PrintGoVersionWarnings() error {
	list, err := this.listOldGoVersionUsage()
	if err != nil {
		return err
	}
	if len(list) > 0 {
		fmt.Println("\n\nthe following repositories use a go version != ", normalizeGoVersion(runtime.Version()))
	}
	slices.SortFunc(list, func(a, b VersionUsageRef) int {
		return strings.Compare(a.Version, b.Version)
	})
	for _, e := range list {
		fmt.Println(e.Version, e.Name)
	}
	return nil
}

func (this *Parsed) listOldGoVersionUsage() (result []VersionUsageRef, err error) {
	current := normalizeGoVersion(runtime.Version())
	for name, mod := range this.Modules {
		if mod.Go != nil {
			modGoVersion := normalizeGoVersion(mod.Go.Version)
			if modGoVersion != current {
				result = append(result, VersionUsageRef{
					Name:    name,
					Version: modGoVersion,
				})
			}
		} else {
			result = append(result, VersionUsageRef{
				Name:    name,
				Version: "missing",
			})
		}
	}
	return result, nil
}

func normalizeGoVersion(version string) string {
	version = strings.ReplaceAll(version, " ", "")
	version = strings.TrimPrefix(version, "go")
	parts := strings.Split(version, ".")
	if len(parts) > 2 {
		version = strings.Join(parts[:2], ".")
	}
	return version
}
