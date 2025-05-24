package config

type NatsConfig struct {
	Host     string `envconfig:"NATS_HOST" validate:"required"`
	Port     int    `envconfig:"NATS_PORT" validate:"required"`
}