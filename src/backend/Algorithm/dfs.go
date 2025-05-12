package algorithm

import "tubes2/model"

func dfs(target string, elements map[string]*model.Element, path []string, visited map[string]bool) []string {
	if contains(path, target) {
		return append(path, target)
	}

	for name, el := range elements {
		if visited[name] || contains(path, name) {
			continue
		}
		for _, recipe := range el.Recipes {
			if containsAll(path, recipe) {
				newPath := append([]string{}, path...)
				newPath = append(newPath, name)
				visited[name] = true
				res := dfs(target, elements, newPath, visited)
				if res != nil {
					return res
				}
				visited[name] = false
			}
		}
	}
	return nil
}
