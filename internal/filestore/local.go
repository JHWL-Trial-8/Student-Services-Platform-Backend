package filestore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Store interface {
	PathFor(hashHex string) (absPath string, objectKey string)
	Exists(hashHex string) (bool, string, string)
	PutFromFile(hashHex, srcPath string) (objectKey string, absPath string, err error)
	Open(hashHex string) (*os.File, string, error)
	Ensure() error
}

type LocalStore struct {
	Root string
}

func NewLocal(root string) *LocalStore {
	return &LocalStore{Root: root}
}

func (s *LocalStore) Ensure() error {
	return os.MkdirAll(filepath.Join(s.Root, "images"), 0o755)
}

func (s *LocalStore) PathFor(hashHex string) (absPath string, objectKey string) {
	sub := "images"
	dir := filepath.Join(sub, hashHex[:2])
	objectKey = filepath.ToSlash(filepath.Join(dir, hashHex))
	absPath = filepath.Join(s.Root, objectKey)
	return
}

func (s *LocalStore) Exists(hashHex string) (bool, string, string) {
	abs, key := s.path(hashHex)
	if _, err := os.Stat(abs); err == nil {
		return true, abs, key
	}
	return false, abs, key
}

func (s *LocalStore) path(hashHex string) (abs string, key string) {
	absPath, objectKey := s.PathFor(hashHex)
	return absPath, objectKey
}

func (s *LocalStore) PutFromFile(hashHex, srcPath string) (string, string, error) {
	abs, key := s.path(hashHex)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", "", err
	}
	// 类原子操作：如果可能，使用 rename；否则回退到复制
	if err := os.Rename(srcPath, abs); err != nil {
		// 跨设备重命名？执行复制
		in, err2 := os.Open(srcPath)
		if err2 != nil {
			return "", "", err2
		}
		defer in.Close()
		out, err2 := os.Create(abs)
		if err2 != nil {
			return "", "", err2
		}
		_, cpErr := io.Copy(out, in)
		closeErr := out.Close()
		_ = os.Remove(srcPath)
		if cpErr != nil {
			return "", "", cpErr
		}
		if closeErr != nil {
			return "", "", closeErr
		}
	}
	return key, abs, nil
}

func (s *LocalStore) Open(hashHex string) (*os.File, string, error) {
	abs, _ := s.path(hashHex)
	f, err := os.Open(abs)
	if err != nil {
		return nil, "", fmt.Errorf("打开文件存储对象失败: %w", err)
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, "", err
	}
	return f, stat.Name(), nil
}