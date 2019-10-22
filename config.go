package ignition

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"

	"os"
	"time"
)

type Config struct {
	mut          sync.Mutex
	fileLoadedAt map[string]int64
}

// Check if the configuration file need reload.
// If the file have been loaded and be modified since then, return true
func (c *Config) NeedReload(file string) bool {
	if c.fileLoadedAt == nil {
		return true
	}
	if loadedAt, ok := c.fileLoadedAt[file]; ok {
		modAt, _ := getFileModAt(file)
		return modAt >= loadedAt
	}

	return true
}

// Load YAML file into configuration variable
func (c *Config) Load(file string, conf interface{}) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.fileLoadedAt == nil {
		c.fileLoadedAt = map[string]int64{}
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	c.fileLoadedAt[file] = time.Now().Unix()
	return yaml.Unmarshal(content, conf)
}

func getFileModAt(file string) (int64, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return 0, err
	}

	return fi.ModTime().Unix(), nil
}
