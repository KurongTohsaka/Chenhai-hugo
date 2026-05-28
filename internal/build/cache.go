package build

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// BuildCache stores file hashes for incremental builds.
type BuildCache struct {
	Files  map[string]string `json:"files"`  // path -> SHA256 hash
	Config string            `json:"config"` // config.yaml hash
	path   string            // cache file path on disk
}

// loadCache loads the build cache from <root>/.chenhai-cache.json.
// If the file doesn't exist or is invalid, returns an empty cache.
func loadCache(root string) (*BuildCache, error) {
	path := filepath.Join(root, ".chenhai-cache.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return &BuildCache{Files: make(map[string]string), path: path}, nil
	}
	var c BuildCache
	if err := json.Unmarshal(data, &c); err != nil {
		return &BuildCache{Files: make(map[string]string), path: path}, nil
	}
	c.path = path
	if c.Files == nil {
		c.Files = make(map[string]string)
	}
	return &c, nil
}

// save writes the cache to disk.
func (c *BuildCache) save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

// hashFile computes SHA256 hash of a file.
func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

// isChanged reports whether a file has changed since it was cached.
// New files (not in cache) are always reported as changed.
// Unreadable files are treated as changed.
func (c *BuildCache) isChanged(path string) (bool, error) {
	hash, err := hashFile(path)
	if err != nil {
		return true, nil // can't read = treat as changed
	}
	oldHash, ok := c.Files[path]
	if !ok {
		return true, nil // new file
	}
	return hash != oldHash, nil
}

// updateFile computes the hash for a file and stores it in the cache.
func (c *BuildCache) updateFile(path string) error {
	hash, err := hashFile(path)
	if err != nil {
		return err
	}
	c.Files[path] = hash
	return nil
}

// deleteFile removes a file entry from the cache.
func (c *BuildCache) deleteFile(path string) {
	delete(c.Files, path)
}

// updateConfig computes and stores the config.yaml hash.
func (c *BuildCache) updateConfig(configPath string) error {
	hash, err := hashFile(configPath)
	if err != nil {
		return err
	}
	c.Config = hash
	return nil
}

// configChanged reports whether config.yaml has changed.
func (c *BuildCache) configChanged(configPath string) bool {
	if c.Config == "" {
		return true // first build
	}
	hash, err := hashFile(configPath)
	if err != nil {
		return true // can't read = treat as changed
	}
	return hash != c.Config
}
