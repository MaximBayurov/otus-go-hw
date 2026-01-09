package sqlstorage

import (
	"fmt"
	"strings"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
)

// makeDsnFromConfig создает строку подключения к базе из настроек.
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
		var format string
		switch value.(type) {
		case int:
			format = "%s=%d "
		case string:
			format = "%s=%s "
		default:
			continue
		}
		dsn.WriteString(fmt.Sprintf(format, key, value))
		dsn.WriteString(" ")
	}

	return strings.TrimRight(dsn.String(), " ")
}
