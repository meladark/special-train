package configs

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port      string
	RedisAddr string
	CLogin    int
	CPass     int
	CIP       int
	RLogin    int
	RPass     int
	RIP       int
}

func LoadConfig() Config {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("REDIS_ADDR", "127.0.0.1:6379")

	viper.SetDefault("CAPACITY_LOGIN", 10)
	viper.SetDefault("CAPACITY_PASS", 100)
	viper.SetDefault("CAPACITY_IP", 1000)

	viper.SetDefault("REFILL_LOGIN", 10)
	viper.SetDefault("REFILL_PASS", 100)
	viper.SetDefault("REFILL_IP", 1000)
	viper.AutomaticEnv()

	cfg := Config{
		Port:      viper.GetString("PORT"),
		RedisAddr: viper.GetString("REDIS_ADDR"),

		CLogin: viper.GetInt("CAPACITY_LOGIN"),
		CPass:  viper.GetInt("CAPACITY_PASS"),
		CIP:    viper.GetInt("CAPACITY_IP"),

		RLogin: viper.GetInt("REFILL_LOGIN"),
		RPass:  viper.GetInt("REFILL_PASS"),
		RIP:    viper.GetInt("REFILL_IP"),
	}
	cfg.prettyPrint()
	return cfg
}

func (c Config) prettyPrint() {
	border := strings.Repeat("=", 40)
	log.Println(border)
	log.Println("Service configuration:")
	log.Printf("  Port:            %s\n", c.Port)
	log.Printf("  Redis Addr:      %s\n", c.RedisAddr)
	log.Println("  --- Buckets ---")
	log.Printf("  Login:           capacity=%d refill/min=%d\n", c.CLogin, c.RLogin)
	log.Printf("  Password:        capacity=%d refill/min=%d\n", c.CPass, c.RPass)
	log.Printf("  IP:              capacity=%d refill/min=%d\n", c.CIP, c.RIP)
	log.Println(border)
}
