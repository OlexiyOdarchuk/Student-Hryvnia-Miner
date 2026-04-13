package app

import (
	"context"

	"shminer/backend/config"
	"shminer/backend/types"
)

type Backend interface {
	StartApp(func(types.LogEntry))
	StartMining(context.Context)
	StopMining()
	IsStorageInitialized() bool
	InitStorage(string) error
	UnlockStorage(string) error
	GetDashboardData() types.DashboardData
	GetWallets() []string
	AddWallet(string, string, string) error
	ImportWalletJSON(string) error
	GetWalletJSONSecure(string, string) (string, error)
	DeleteWallet(string, string) error
	RenameWallet(string, string) error
	UpdateWalletKey(string, string, string) error
	GetWalletKey(string, string) (string, error)
	ToggleWallet(string) bool
	SetGlobalMining(bool) error
	GetConfig() config.AppConfig
	UpdateConfig(config.AppConfig, string) error
	ChangePassword(string, string) error
	SendTransaction(from, to, password string, amount int) error
	SendMessageToDeveloper(contact, message string)
	TryAutoLogin() (bool, error)
	GetConfigFilePath() string
}

func Init() Backend {
	return New()
}
