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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TagResult struct {
	Results []TagStruct `json:"results"`
}

type TagStruct struct {
	Name string `json:"name"`
}

func getGolangTagsFromDockerhub() (tags []string, err error) {
	resp, err := http.Get("https://hub.docker.com/v2/repositories/library/golang/tags")
	if err != nil {
		return tags, err
	}
	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return tags, fmt.Errorf("unexpected statuscode %v %v", resp.StatusCode, string(payload))
	}
	temp := TagResult{}
	err = json.NewDecoder(resp.Body).Decode(&temp)
	if err != nil {
		return tags, err
	}
	for _, tag := range temp.Results {
		tags = append(tags, tag.Name)
	}
	return tags, nil
}
