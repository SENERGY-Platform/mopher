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
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"math/rand"
	"slices"
	"strings"
)

type TextNode struct {
	Id   int64
	Text string
}

func (this *TextNode) ID() int64 {
	return this.Id
}

func NewTextNode(text string) *TextNode {
	return &TextNode{
		Id:   rand.Int63(),
		Text: text,
	}
}

func (this *Parsed) GetRecommendedUpdateOrder() (result []string, err error) {
	textToNodes := map[string]*TextNode{}

	g := simple.NewDirectedGraph()
	for called, callers := range this.Inverse {
		if _, ok := this.Modules[called]; ok {
			for _, caller := range callers {
				if _, ok := this.Modules[caller.UserModule]; ok {
					callerNode, ok := textToNodes[caller.UserModule]
					if !ok {
						callerNode = NewTextNode(caller.UserModule)
						textToNodes[caller.UserModule] = callerNode
					}
					calledNode, ok := textToNodes[called]
					if !ok {
						calledNode = NewTextNode(called)
						textToNodes[called] = calledNode
					}
					if g.Node(callerNode.Id) == nil {
						g.AddNode(callerNode)
					}
					if g.Node(calledNode.Id) == nil {
						g.AddNode(calledNode)
					}
					g.SetEdge(simple.Edge{T: callerNode, F: calledNode})
				}
			}
		}
	}
	nodes, err := topo.SortStabilized(g, func(nodes []graph.Node) {
		slices.SortFunc(nodes, func(a, b graph.Node) int {
			return strings.Compare(a.(*TextNode).Text, b.(*TextNode).Text)
		})
	})
	if err != nil {
		return result, err
	}
	for _, n := range nodes {
		result = append(result, n.(*TextNode).Text)
	}
	return result, nil
}

func (this *Parsed) toBeUpdated(filter map[string]bool, e string) bool {
	if filter[e] {
		return true
	}
	module, ok := this.Modules[e]
	if ok {
		for _, req := range module.Require {
			if this.toBeUpdated(filter, req.Mod.Path) {
				return true
			}
		}
	}
	return false
}
