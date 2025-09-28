package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress string `env:"RUN_ADDRESS"`
	DatabaseURL string `env:"DATABASE_URL"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func ParseConfig() *Config {
	var cfg Config

	cfg.RunAddress = ""
	cfg.DatabaseURL = ""
	cfg.AccrualSystemAddress = ""

	_ = env.Parse(&cfg)

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "service address")
	flag.StringVar(&cfg.DatabaseURL, "d", cfg.DatabaseURL, "database url")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")
	return &cfg
}