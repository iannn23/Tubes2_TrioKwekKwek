package algorithm

import "tubes2/model"

type PathNode struct {
	Name string
	Path []string
}

func bfs(target string, elements map[string]*model.Element) []string {
	queue := []PathNode{}
	visited := map[string]bool{}
	base := []string{"air", "earth", "fire", "water"}

	for _, b := range base {
		queue = append(queue, PathNode{Name: b, Path: []string{b}})
		visited[b] = true
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.Name == target {
			return curr.Path
		}

		for name, el := range elements {
			for _, recipe := range el.Recipes {
				if containsAll(curr.Path, recipe) && !visited[name] {
					visited[name] = true
					queue = append(queue, PathNode{Name: name, Path: append(append([]string{}, curr.Path...), name)})
				}
			}
		}
	}
	return nil
}
