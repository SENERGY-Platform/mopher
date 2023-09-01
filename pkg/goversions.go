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
	"golang.org/x/mod/semver"
	"log/slog"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
)

func (this *Parsed) PrintGoVersionWarnings() (deprecated []string, err error) {
	list, err := this.listOldGoVersionUsage()
	if err != nil {
		return deprecated, err
	}
	if len(list) > 0 {
		fmt.Println("\n\nthe following repositories use a go version !=", normalizeGoVersion(getLatestGoVersion()))
	}
	slices.SortFunc(list, func(a, b VersionUsageRef) int {
		result := strings.Compare(a.Version, b.Version)
		if result == 0 {
			result = strings.Compare(a.Name, b.Name)
		}
		return result
	})
	for _, e := range list {
		fmt.Println(e.Version, e.Name)
		deprecated = append(deprecated, e.Name)
	}
	return deprecated, nil
}

func (this *Parsed) listOldGoVersionUsage() (result []VersionUsageRef, err error) {
	current := normalizeGoVersion(getLatestGoVersion())

	//make result deterministic by sorting the keys
	keys := []string{}
	for key, _ := range this.Latest {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, name := range keys {
		mod := this.Modules[name]
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

func getLatestGoVersion() string {
	buildVersion := runtime.Version()
	tags, err := getGolangTagsFromDockerhub()
	if err != nil {
		slog.Debug(fmt.Sprint("unable to load tags from dockerhub:", err))
		slog.Debug("fallback to mopher build go version")
		return runtime.Version()
	}
	for _, tag := range tags {
		if isValidSemanticVersion(tag) && semver.Compare(buildVersion, tag) < 1 {
			return tag
		}
	}
	slog.Debug("no go build tag found at dockerhub. fallback to mopher build go version")
	return runtime.Version()
}

func isValidSemanticVersion(version string) bool {
	re := regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
	return re.MatchString(version)
}
