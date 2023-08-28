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
	"golang.org/x/mod/modfile"
	"strings"
)

func LoadOrg(org string) (parsed *Parsed, err error) {
	parsed = &Parsed{
		Modules: map[string]*modfile.File{},
		Inverse: map[string][]InverseIndexModRef{},
		Latest:  map[string]LatestCommitInfo{},
	}
	parsed.Repos, err = loadOrgRepos(org)
	for _, repo := range parsed.Repos {
		if repo.Language != nil && *repo.Language == "Go" && (repo.Archived == nil || *repo.Archived == false) {
			name, module, err := parseRepoModuleFile(repo)
			if err == ErrModfileNotFound {
				continue
			}
			if err != nil {
				return parsed, err
			}
			parsed.Modules[name] = module
			parsed.Latest[name], err = getLatestInfo(repo)
			if err != nil {
				return parsed, err
			}
		}
	}
	for name, module := range parsed.Modules {
		for _, req := range module.Require {
			version, semantic := normalizeGoDependencyVersion(req.Mod.Version)
			parsed.Inverse[req.Mod.Path] = append(parsed.Inverse[req.Mod.Path], InverseIndexModRef{
				UsesVersion:     version,
				UserModule:      name,
				SemanticVersion: semantic,
			})
		}
	}

	return parsed, nil
}

func normalizeGoDependencyVersion(version string) (string, bool) {
	parts := strings.Split(version, "-")
	switch len(parts) {
	case 1:
		return version, true
	default:
		return parts[len(parts)-1], false
	}
}
