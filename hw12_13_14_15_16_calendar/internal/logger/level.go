package logger

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelMap = map[string]LogLevel{
	"DEBUG": DEBUG,
	"INFO":  INFO,
	"WARN":  WARN,
	"ERROR": ERROR,
	"FATAL": FATAL,
}

// String возвращает строковое представление
func (l LogLevel) String() string {
	for s, level := range levelMap {
		if level == l {
			return s
		}
	}
	return ""
}

// ParseStatus преобразует строку в Status
func ParseStatus(level string) LogLevel {
	result := DEBUG

	if result, ok := levelMap[level]; ok {
		return result
	}
	return result
}
