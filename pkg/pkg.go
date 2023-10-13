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
	"io"
	"log"
)

func Mopher(writer io.Writer, org string, maxConn int, graph string, verbose bool, dep string, warnUnsyncDev bool) {
	if org == "" {
		log.Fatal("missing org input")
		return
	}

	parsed, err := LoadOrg(org, maxConn)
	if err != nil {
		log.Fatal(err)
		return
	}

	parsed.SetOutput(writer)

	if graph != "" {
		err = parsed.StoreGraph(graph, verbose)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	if dep != "" {
		err = parsed.PrintDependents(dep)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	err = parsed.PrintWarnings(warnUnsyncDev)
	if err != nil {
		log.Fatal(err)
		return
	}
}
