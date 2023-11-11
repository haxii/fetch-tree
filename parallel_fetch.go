package fetch_tree

import (
	"fmt"
	"github.com/haxii/task"
	"sync"
)

type Do func(nodeID string, depth int) ([]NodeInfo, error)

func DoParallel(fetch Do, treeName, rootNodeID string, maxDepth int, thread int) (Tree, error) {
	cache, cacheErr := newFetchCache()
	if cacheErr != nil {
		return nil, cacheErr
	}
	defer cache.close()
	fetchWithCache := func(_nodeID string, _depth int) ([]NodeInfo, error) {
		if info, cached := cache.get(treeName, _nodeID); cached {
			fmt.Printf("hit cache %s:%s\n", treeName, _nodeID)
			return info, nil
		}
		info, err := fetch(_nodeID, _depth)
		if err != nil {
			return nil, err
		}
		return info, cache.save(treeName, _nodeID, info)
	}
	if cacheErr != nil {
		return nil, cacheErr
	}
	if maxDepth > 4 {
		maxDepth = 4
	}
	rootNodeList, err := fetchWithCache(rootNodeID, 1)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0)
	rootsNodes := make(NodeMap)
	for _, info := range rootNodeList {
		rootsNodes[info.ID] = info.Name
		if !info.IsLeaf {
			keys = append(keys, info.ID)
		}
	}

	nodes := make(Tree)
	nodes[rootNodeID] = rootsNodes
	if maxDepth <= 1 {
		return nodes, nil
	}

	lock := sync.Mutex{}
	addNodes := func(id string, l []NodeInfo) {
		if len(l) == 0 {
			return
		}
		lock.Lock()
		defer lock.Unlock()
		nodes[id] = make(NodeMap)
		for _, info := range l {
			nodes[id][info.ID] = info.Name
		}
	}
	var fetchAll func(currentNodeID string, currentDepth int) error
	fetchAll = func(currentNodeID string, currentDepth int) error {
		currentSubList, fetchErr := fetchWithCache(currentNodeID, currentDepth)
		if fetchErr != nil {
			return fetchErr
		}
		addNodes(currentNodeID, currentSubList)
		if currentDepth >= maxDepth {
			return nil
		}
		for _, info := range currentSubList {
			if !info.IsLeaf {
				if fetchErr = fetchAll(info.ID, currentDepth+1); fetchErr != nil {
					return fetchErr
				}
			}
		}
		return nil
	}
	err = task.Execute(keys, thread, func(currentNodeID string) error {
		return fetchAll(currentNodeID, 2)
	})
	return nodes, err
}
