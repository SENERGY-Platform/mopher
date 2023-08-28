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
	"io"
	"net/http"
	"path"
)

var ErrModfileNotFound = errors.New("no go.mod file in root of default branch of the repository")

func parseRepoModuleFile(repo *github.Repository) (moduleName string, module *modfile.File, err error) {
	fullName := ""
	if repo.FullName != nil {
		fullName = *repo.FullName
	} else {
		return moduleName, module, errors.New("missing repo FullName")
	}
	defaultBranch := "master"
	if repo.DefaultBranch != nil {
		defaultBranch = *repo.DefaultBranch
	}
	resp, err := http.Get(GithubRawUrl + path.Join(fullName, defaultBranch, "go.mod"))
	if err != nil {
		return moduleName, module, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return moduleName, module, ErrModfileNotFound
	}
	file, err := io.ReadAll(resp.Body)
	if err != nil {
		return moduleName, module, err
	}
	module, err = modfile.ParseLax("go.mod", file, nil)
	if err != nil {
		return moduleName, module, err
	}
	moduleName = module.Module.Mod.Path
	return moduleName, module, nil
}
