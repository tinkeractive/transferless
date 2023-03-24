package configuration

import (
	"os"
	"path/filepath"
)

type Manager struct{}

func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

func (m *Manager) SaveConfig(cfg string) error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "/.config/rclone")
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return err
	}
	configPath := filepath.Join(path, "/rclone.conf")
	return os.WriteFile(configPath, []byte(cfg), 0666)
}
