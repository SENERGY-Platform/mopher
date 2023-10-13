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
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v54/github"
	"golang.org/x/mod/semver"
	"log/slog"
)

func getLatestInfo(repo *github.Repository) (result LatestCommitInfo, err error) {
	remoteUrl := repo.GetSSHURL()
	slog.Debug("git ls-remote " + remoteUrl)
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteUrl},
	})
	refs, err := rem.List(&git.ListOptions{
		Timeout: 30,
	})
	if err != nil {
		return result, err
	}

	defaultRefName := plumbing.Master
	//recheck head, in case 'main' branch is used
	for _, ref := range refs {
		if ref.Name() == plumbing.HEAD {
			defaultRefName = ref.Target()
			break
		}
	}
	//find latest tag
	latestTag := ""
	for _, ref := range refs {
		if ref.Name().IsTag() {
			refVersion := ref.Name().Short()
			if latestTag == "" || semver.Compare(latestTag, refVersion) < 1 {
				latestTag = refVersion
			}
		}
	}

	result = LatestCommitInfo{
		LatestTag: latestTag,
	}

	var devRefName plumbing.ReferenceName = "refs/heads/dev"
	for _, ref := range refs {
		if ref.Name() == defaultRefName {
			result.MainHash = shorHash(ref.Hash().String())
		}
		if ref.Name() == devRefName {
			result.DevHash = shorHash(ref.Hash().String())
		}
	}

	if result.MainHash == "" {
		return result, errors.New("no HEAD found")
	}

	return result, nil
}

func shorHash(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}
