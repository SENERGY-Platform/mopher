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
	"strings"
)

func (this *Parsed) generatePlantuml(verbose bool) (string, error) {
	header := "@startuml\n!pragma layout elk\n\n"
	footer := "\n\n@enduml"
	lines := []string{}

	orgRepos := map[string]bool{}
	for name, _ := range this.Modules {
		orgRepos[name] = true
	}
	for called, callers := range this.Inverse {
		if verbose || orgRepos[called] {
			lines = append(lines, fmt.Sprintf("\n'dependent on %v", called))
			lines = append(lines, fmt.Sprintf("[%v]", called))
			for _, caller := range callers {
				lines = append(lines, fmt.Sprintf("[%v] --> [%v]: %v", caller.UserModule, called, caller.UsesVersion))
			}
		}
	}
	return header + strings.Join(lines, "\n") + footer, nil
}
