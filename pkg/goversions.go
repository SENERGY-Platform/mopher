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
	checkedGoVersion := normalizeGoVersion(getLatestGoVersion())
	list, err := this.listOldGoVersionUsage(checkedGoVersion)
	if err != nil {
		return deprecated, err
	}
	if len(list) > 0 {
		_, err = fmt.Fprintln(this.output, "\n\nthe following repositories use a go version !=", checkedGoVersion)
		if err != nil {
			return deprecated, err
		}
	}
	slices.SortFunc(list, func(a, b VersionUsageRef) int {
		result := strings.Compare(a.Version, b.Version)
		if result == 0 {
			result = strings.Compare(a.Name, b.Name)
		}
		return result
	})
	for _, e := range list {
		_, err = fmt.Fprintln(this.output, e.Version, e.Name)
		if err != nil {
			return deprecated, err
		}
		deprecated = append(deprecated, e.Name)
	}
	return deprecated, nil
}

func (this *Parsed) listOldGoVersionUsage(checkedGoVersion string) (result []VersionUsageRef, err error) {
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
			if modGoVersion != checkedGoVersion {
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
	buildVersion := normalizeGoVersion(runtime.Version())
	tags, err := getGolangTagsFromDockerhub()
	if err != nil {
		slog.Debug("unable to load tags from dockerhub:", "err", err)
		slog.Debug("fallback to mopher build go version")
		return buildVersion
	}
	tags = append(tags, buildVersion)
	tags = dockerhubTagCleanup(tags)
	slices.SortFunc(tags, func(a, b string) int {
		return semver.Compare(ensureSemverComparable(b), ensureSemverComparable(a))
	})
	slog.Debug("getLatestGoVersion()", "used-tag", tags[0], "known-tags", tags)
	return tags[0]
}

func dockerhubTagCleanup(tags []string) (result []string) {
	for _, tag := range tags {
		trimmed := strings.Split(tag, "-")[0]
		if isValidSemanticVersion(trimmed) && !slices.Contains(result, trimmed) {
			result = append(result, trimmed)
		}
	}
	return result
}

func ensureSemverComparable(version string) string {
	if version == "" || version[0] != 'v' {
		version = "v" + version
	}
	return version
}

func isValidSemanticVersion(version string) bool {
	re := regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
	return re.MatchString(version)
}
