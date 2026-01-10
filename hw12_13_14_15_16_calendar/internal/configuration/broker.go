package configuration

import "fmt"

type BrokerConf struct {
	User              string `mapstructure:"user"`
	Password          string `mapstructure:"password"`
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	ExchangeName      string `mapstructure:"exchange_name"`
	ExchangeType      string `mapstructure:"exchange_type"`
	QueueName         string `mapstructure:"queue_name"`
	RoutingKey        string `mapstructure:"routing_key"`
	Durable           bool   `mapstructure:"durable"`
	AutoDelete        bool   `mapstructure:"auto_delete"`
	Exclusive         bool   `mapstructure:"exclusive"`
	NoWait            bool   `mapstructure:"no_wait"`
	ReconnectInterval int64  `mapstructure:"reconnect_interval"`
}

func (r BrokerConf) URL() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		r.User,
		r.Password,
		r.Host,
		r.Port,
	)
}
