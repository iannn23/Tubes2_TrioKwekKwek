package algorithm

import "tubes2/model"

func Search(target, algo string, multiple bool, max int, elements map[string]*model.Element) interface{} {
	if multiple {
		return FindMultipleRecipes(target, elements, max)
	}
	if algo == "DFS" {
		return dfs(target, elements, []string{}, map[string]bool{})
	}
	return bfs(target, elements)
}
