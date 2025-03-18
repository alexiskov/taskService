package confreader

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type (
	Config struct {
		DB     DataBase
		WebSRV WebServer
	}

	DataBase struct {
		Ip       string
		Port     string
		Username string
		Password string
		DBname   string
	}

	WebServer struct {
		Port string
	}
)

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("config load error: %w", err)
	}
	return &Config{DB: DataBase{Ip: os.Getenv("DB_IP_ADDRESS"), Port: os.Getenv("DB_PORT"), Username: os.Getenv("DB_USERNAME"), Password: os.Getenv("DB_PASSWORD"), DBname: os.Getenv("DB_NAME")}, WebSRV: WebServer{Port: os.Getenv("WEB_PORT")}}, nil
}
