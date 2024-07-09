/*
 * Copyright 2024 InfAI (CC SES)
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
	"golang.org/x/mod/modfile"
	"strings"
)

type UpdateModeCommand struct {
	Cmd  string
	Args []string
}

func RunUpdateMode(mod *modfile.File, internal bool) (commands []UpdateModeCommand, err error) {
	org := getOrgOfGithubPath(mod.Module.Mod.Path)
	for _, req := range mod.Require {
		if getOrgOfGithubPath(req.Mod.Path) == org {
			version, semantic := normalizeGoDependencyVersion(req.Mod.Version)
			latest, err := getLatestInfoFromUrl("https://" + req.Mod.Path + ".git")
			if err != nil {
				return nil, err
			}
			latestVersionStr := latest.MainHash
			if semantic {
				latestVersionStr = latest.LatestTag
			}
			if version != latestVersionStr {
				commands = append(commands, UpdateModeCommand{
					Cmd:  "go",
					Args: []string{"get", fmt.Sprintf("%v@%v", req.Mod.Path, latestVersionStr)},
				})
			}
		}
	}
	if internal {
		commands = append(commands,
			UpdateModeCommand{
				Cmd:  "go",
				Args: []string{"mod", "tidy"},
			},
		)
	} else {
		commands = append(commands,
			UpdateModeCommand{
				Cmd:  "go",
				Args: []string{"get", "-u", "-t", "./..."},
			},
			UpdateModeCommand{
				Cmd:  "go",
				Args: []string{"mod", "tidy"},
			},
		)
	}
	return commands, nil
}

func getOrgOfGithubPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 3 {
		return strings.Join(parts[:2], "/")
	}
	return path
}
