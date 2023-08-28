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
	"context"
	"fmt"
	"github.com/google/go-github/v54/github"
	"log/slog"
)

func loadOrgRepos(org string) (result []*github.Repository, err error) {
	client := github.NewClient(nil)
	options := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}
	for {
		slog.Debug(fmt.Sprintf("request github %v %v", options.ListOptions.PerPage, options.ListOptions.Page))
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), org, options)
		if err != nil {
			return result, err
		}
		if resp.NextPage == 0 {
			break
		} else {
			options.ListOptions.Page = resp.NextPage
		}
		result = append(result, repos...)
	}
	return result, nil
}
