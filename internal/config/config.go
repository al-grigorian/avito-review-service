package config

import (
    "os"
)

type Config struct {
    DBConn string
}

func Load() *Config {
    return &Config{
        DBConn: os.Getenv("DB_CONN"),
    }
}