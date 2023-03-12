package configuration

import (
	"context"

	"github.com/rclone/rclone/fs/config/configfile"
)

type Provider interface {
	GetConfig() (string, error)
}

func LoadConfig(ctx context.Context, cfg string) error {
	configManager, err := NewManager()
	if err != nil {
		return err
	}
	err = configManager.SaveConfig(cfg)
	if err != nil {
		return err
	}
	configfile.Install()
	return nil
}
