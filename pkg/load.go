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
	"errors"
	"github.com/google/go-github/v54/github"
	"golang.org/x/mod/modfile"
	"strings"
	"sync"
)

func LoadOrg(org string, maxConn int) (parsed *Parsed, err error) {
	parsed = &Parsed{
		org:     org,
		Modules: map[string]*modfile.File{},
		Inverse: map[string][]InverseIndexModRef{},
		Latest:  map[string]LatestCommitInfo{},
	}
	parsed.Repos, err = loadOrgRepos(org)
	if err != nil {
		return parsed, err
	}
	parsed.Modules, parsed.Latest, err = getRepoInfos(parsed.Repos, maxConn)
	if err != nil {
		return parsed, err
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

func getRepoInfos(repos []*github.Repository, maxConn int) (modules map[string]*modfile.File, latestInfo map[string]LatestCommitInfo, err error) {
	modules = map[string]*modfile.File{}
	latestInfo = map[string]LatestCommitInfo{}
	mux := sync.Mutex{}
	wg := sync.WaitGroup{}
	asyncErrors := []error{}
	limit := make(chan bool, maxConn)
	for _, repo := range repos {
		if repo.Language != nil && *repo.Language == "Go" && (repo.Archived == nil || *repo.Archived == false) {
			wg.Add(1)
			go func(r *github.Repository) {
				defer wg.Done()
				limit <- true
				defer func() {
					<-limit
				}()
				name, module, err := parseRepoModuleFile(r)
				if err == ErrModfileNotFound {
					return
				}
				if err != nil {
					asyncErrors = append(asyncErrors, err)
					return
				}
				latest, err := getLatestInfo(r)
				mux.Lock()
				defer mux.Unlock()
				if err != nil {
					asyncErrors = append(asyncErrors, err)
					return
				}
				modules[name] = module
				latestInfo[name] = latest
			}(repo)
		}
	}
	wg.Wait()
	if len(asyncErrors) > 0 {
		return modules, latestInfo, errors.Join(asyncErrors...)
	}
	return modules, latestInfo, nil
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
