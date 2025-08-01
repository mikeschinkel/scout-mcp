package scoutcfg_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/mikeschinkel/scout-mcp/scoutcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var logger *slog.Logger

type testData struct {
	Name string
	Age  int
}

func TestFileStore_SaveLoadExists(t *testing.T) {
	var err error
	dir := filepath.Join(os.TempDir(), "gmcfg-test-"+uuid.NewString())
	t.Cleanup(func() { must(os.RemoveAll(dir)) })

	s := scoutcfg.NewFileStore("test-app")
	s.SetBaseDir(dir)

	data := testData{Name: "Alice", Age: 42}
	filename := "config/testdata.json"

	err = s.Save(filename, &data)
	require.NoError(t, err)

	exists := s.Exists(filename)
	assert.True(t, exists)

	var loaded testData
	err = s.Load(filename, &loaded)
	require.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestFileStore_Append(t *testing.T) {
	var err error
	dir := filepath.Join(os.TempDir(), "gmcfg-append-"+uuid.NewString())
	t.Cleanup(func() { must(os.RemoveAll(dir)) })

	s := scoutcfg.NewFileStore("test-app")
	s.SetBaseDir(dir)

	filename := "log/output.log"
	msg1 := []byte("first line\n")
	msg2 := []byte("second line\n")

	err = s.Append(filename, msg1)
	require.NoError(t, err)

	err = s.Append(filename, msg2)
	require.NoError(t, err)

	path := filepath.Join(dir, filename)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(msg1)+string(msg2), string(content))
}

func TestFileStore_LoadNonexistent(t *testing.T) {
	var err error

	s := scoutcfg.NewFileStore("test-app")
	s.SetBaseDir(t.TempDir())

	err = s.Load("does-not-exist.json", &testData{})
	assert.Error(t, err)
}

func TestFileStore_SaveInvalidJSON(t *testing.T) {
	var err error

	s := scoutcfg.NewFileStore("test-app")
	s.SetBaseDir(t.TempDir())

	ch := make(chan int) // non-serializable
	err = s.Save("bad.json", ch)
	assert.Error(t, err)
}

func TestFileStore_ConfigDir(t *testing.T) {
	s := scoutcfg.NewFileStore("test-app")
	dir := t.TempDir()
	s.SetBaseDir(dir)

	cfgDir, err := s.ConfigDir()
	assert.NoError(t, err)
	assert.Equal(t, dir, cfgDir)
}

func must(err error) {
	if err != nil {
		logger.Error(err.Error())
	}
}
