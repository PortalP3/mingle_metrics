package config

import (
	"testing"

	"os"

	"io/ioutil"
	"path/filepath"

	"github.com/stretchr/testify/assert"
)

var validJson = `
{
	"Login": "my_login",
	"ProjectID": "my_id",
	"Secret": "my_secret",
	"Endpoint": "my_endpoint"
}
`

func TestFileNotFound(t *testing.T) {
	_, err := Load("testdata/invalid_config.json")
	assert.True(t, os.IsNotExist(err), "")
}

func TestJsonParsingConfig(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "config")
	f.WriteString(validJson)
	assert.NoError(t, err)
	actual, err := Load(filepath.Join(f.Name()))
	assert.Equal(t, "my_login", actual.Login)
	assert.Equal(t, "my_id", actual.ProjectID)
	assert.Equal(t, "my_secret", actual.Secret)
	assert.Equal(t, "my_endpoint", actual.Endpoint)
}
