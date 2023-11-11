package fetch_tree

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type cacheInfo struct {
	TreeName string
	NodeID   string
	Children []NodeInfo
}

type fetchCache struct {
	cacheMap  map[string]map[string]cacheInfo
	cacheFile *os.File
}

func newFetchCache() (*fetchCache, error) {
	cachePath := filepath.Join(os.TempDir(), fmt.Sprintf(
		"fetch-cache-%s.jsonl", time.Now().Format(time.DateOnly)))
	fmt.Println("use cache", cachePath)
	f, err := os.OpenFile(cachePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(f)
	m := make(map[string]map[string]cacheInfo)
	for s.Scan() {
		info := cacheInfo{}
		if err = json.Unmarshal(s.Bytes(), &info); err != nil || len(info.TreeName) == 0 {
			fmt.Printf("skipped %s with error: %s", s.Bytes(), err)
			continue
		}
		if len(m[info.TreeName]) == 0 {
			m[info.TreeName] = make(map[string]cacheInfo)
		}
		m[info.TreeName][info.NodeID] = info
	}

	return &fetchCache{
		cacheMap:  m,
		cacheFile: f,
	}, nil
}

func (c fetchCache) close() error {
	return c.cacheFile.Close()
}

func (c fetchCache) get(treeName, nodeID string) ([]NodeInfo, bool) {
	m, exists := c.cacheMap[treeName]
	if !exists {
		return nil, false
	}
	l, ok := m[nodeID]
	return l.Children, ok
}

func (c fetchCache) save(treeName, nodeID string, info []NodeInfo) error {
	if info == nil {
		return nil
	}
	b, err := json.Marshal(&cacheInfo{
		TreeName: treeName, NodeID: nodeID, Children: info})
	if err != nil {
		return err
	}
	_, err = c.cacheFile.Write(append(b, '\n'))
	return err
}
