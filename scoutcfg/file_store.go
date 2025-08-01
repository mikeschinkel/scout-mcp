package scoutcfg

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const ConfigBaseDirName = ".config"

type FileStore struct {
	appName   string
	configDir string
	fs        fs.FS
}

func NewFileStore(appName string) *FileStore {
	return &FileStore{
		appName: appName,
	}
}

func (s *FileStore) ConfigDir() (_ string, err error) {
	var homeDir string
	if s.configDir != "" {
		goto end
	}

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	s.configDir = filepath.Join(homeDir, ConfigBaseDirName, s.appName)

end:
	return s.configDir, err
}

func (s *FileStore) getFS() (_ fs.FS, err error) {
	var dir string

	if s.fs != nil {
		goto end
	}

	dir, err = s.ConfigDir()
	if err != nil {
		goto end
	}

	s.fs = os.DirFS(dir)

end:
	return s.fs, err
}

func (s *FileStore) ensureFilepath(filename string) (fp string, err error) {
	fp, err = s.getFilepath(filename)
	// This is needed in case filename contains a subdirectory, e.g. tokens/token-bill@microsoft.com.json
	err = os.MkdirAll(filepath.Dir(fp), 0755)
	if err != nil {
		goto end
	}
end:
	return fp, err
}

func (s *FileStore) getFilepath(filename string) (fp string, err error) {
	var dir string

	dir, err = s.ConfigDir()
	if err != nil {
		goto end
	}

	if !fs.ValidPath(filename) {
		err = fmt.Errorf("path %s is not valid for use in %s", filename, dir)
		goto end
	}

	fp = filepath.Join(s.configDir, filename)

end:
	return fp, err
}

func (s *FileStore) Save(filename string, data any) (err error) {
	var jsonData []byte
	var file *os.File
	var fullPath string

	ensureLogger()

	jsonData, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		goto end
	}

	fullPath, err = s.ensureFilepath(filename)
	if err != nil {
		goto end
	}

	file, err = os.Create(fullPath)
	if err != nil {
		goto end
	}
	defer mustClose(file)

	_, err = file.Write(jsonData)

end:
	return err
}

func (s *FileStore) Load(filename string, data any) (err error) {
	var jsonData []byte
	var fsys fs.FS

	fsys, err = s.getFS()
	if err != nil {
		goto end
	}

	jsonData, err = fs.ReadFile(fsys, filename)
	if err != nil {
		goto end
	}

	err = json.Unmarshal(jsonData, data)

end:
	return err
}

func (s *FileStore) Append(filename string, content []byte) (err error) {
	var file *os.File
	var fullPath string

	fullPath, err = s.ensureFilepath(filename)
	if err != nil {
		goto end
	}

	file, err = os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		goto end
	}
	defer mustClose(file)

	_, err = file.Write(content)
	if err != nil {
		goto end
	}

	err = file.Sync()

end:
	return err
}

func (s *FileStore) Exists(filename string) (exists bool) {
	fsys, err := s.getFS()
	if err != nil {
		goto end
	}
	_, err = fs.Stat(fsys, filename)
	exists = err == nil

end:
	return exists
}

func (s *FileStore) SetBaseDir(dir string) {
	s.configDir = dir
	s.fs = os.DirFS(dir)
}
