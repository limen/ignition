package ignition

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type myConf struct {
	Locale  string `yml:"locale"`
	AppKey  string `yml:"app_key"`
	AppName string `yml:"app_name"`
}

func TestConfig(t *testing.T) {
	var conf = Config{}
	myconf := &myConf{}
	err := conf.Load("", myconf)
	assert.Error(t, err)

	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)

	filePath := basePath + "/examples/.env.yml"
	assert.True(t, conf.NeedReload(filePath))
	assert.Nil(t, conf.Load(filePath, myconf))

	time.Sleep(time.Second * 2)

	assert.False(t, conf.NeedReload(filePath))
	assert.True(t, len(myconf.Locale) > 0)
	assert.True(t, len(myconf.AppKey) == 0)

	fmt.Println("Sleeping 20 senconds to modify yml file: examples/.env.yml")
	time.Sleep(time.Second * 20)
	fmt.Println("Sleep time over")
	// modify yml file in time
	assert.True(t, conf.NeedReload(filePath))
}
