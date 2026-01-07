package sqlstorage

import (
	"fmt"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"strings"
)

// makeDsnFromConfig создает строку подключения к базе из настроек
func makeDsnFromConfig(configs configuration.DbConf) string {
	var dsn strings.Builder
	dsnParams := map[string]any{
		"host":     configs.Host,
		"port":     configs.Port,
		"dbname":   configs.DbName,
		"user":     configs.User,
		"password": configs.Password,
	}

	for key, value := range dsnParams {
		dsn.WriteString(fmt.Sprintf("%s=%s", key, value))
	}
	return dsn.String()
}
