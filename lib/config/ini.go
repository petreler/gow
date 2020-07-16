package config

import (
	ini "github.com/go-ini/ini"
	"os"
	"strings"
)

const (
	defaultConf     = "conf/app.conf"
	defaultDevCong  = "conf/dev.app.conf"
	defaultProdConf = "conf/prod.app.conf"
)

var (
	cfg      = ini.Empty()
	fileName string
)

// init load current configuration file
func init() {
	runMode := os.Getenv("APP_RUN_MODE")
	if runMode == "" {
		fileName = defaultConf
	}

	if runMode == "dev" {
		fileName = defaultDevCong
	}
	if runMode == "prod" {
		fileName = defaultProdConf
	}

	if fileName == "" {
		fileName = defaultConf
	}

	var err error
	cfg, err = ini.InsensitiveLoad(fileName)
	if err != nil {
		panic("Failed to read configuration file")
	}

}

// DefaultString
func DefaultString(key, def string) string {
	if v := GetString(key); v != "" {
		return v
	}
	return def
}

// GetString
func GetString(key string) string {
	return getKey(key).String()
}

//DefaultInt DefaultInt
func DefaultInt(key string, def int) int {
	if v, err := GetInt(key); err == nil {
		return v
	}
	return def
}

func GetInt(key string) (int, error) {
	return getKey(key).Int()
}

//DefaultInt DefaultInt
func DefaultInt64(key string, def int64) int64 {
	if v, err := GetInt64(key); err == nil {
		return v
	}
	return def
}

func GetInt64(key string) (int64, error) {
	return getKey(key).Int64()
}

//DefaultInt DefaultInt
func DefaultFloat(key string, def float64) float64 {
	if v, err := GetFloat(key); err == nil {
		return v
	}
	return def
}

func GetFloat(key string) (float64, error) {
	return getKey(key).Float64()
}

//GetInt64
func GetBool(key string) (bool, error) {
	return getKey(key).Bool()
}

//DefaultBool DefaultBool
func DefaultBool(key string, def bool) bool {
	if v, err := GetBool(key); err == nil {
		return v
	}
	return def
}

// Keys 获取section下的所有keys
func Keys(section string) []string {
	return cfg.Section(section).KeyStrings()
}

//getKey getKey
func getKey(key string) *ini.Key {
	if key == "" {
		return nil
	}
	sp := strings.Split(key, "::")
	switch len(sp) {
	case 1:
		return cfg.Section("").Key(key)
	case 2:
		return cfg.Section(sp[0]).Key(sp[1])
	default:
		return nil
	}

}
