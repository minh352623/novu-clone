package initialize

import (
	"fmt"
	"os"

	"CONVERDA/global"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func LoadConfig() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	viper := viper.New()
	viper.AddConfigPath("./config/")
	viper.SetConfigName("environment")
	viper.SetConfigType("yaml")

	// read configuration
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("failed to read configuration: %w", err))
	}

	// Substitute environment variables in the YAML file
	// This step ensures that placeholders like ${PORT} are replaced with actual values
	for _, key := range viper.AllKeys() {
		value := viper.GetString(key)
		if value != "" {
			viper.Set(key, os.ExpandEnv(value)) // os.ExpandEnv replaces ${VAR} with the value of VAR
		}
	}

	// configuration with struct
	if err := viper.Unmarshal(&global.Config); err != nil {
		panic(fmt.Errorf("unable to decode configuration: %w", err))
	}
}
