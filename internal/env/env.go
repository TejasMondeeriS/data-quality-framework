package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v2"
)

var config map[string]interface{}

func init() {
	config = make(map[string]interface{})

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not get current working directory, proceeding with env variables only")
		return
	}

	configFilePath := filepath.Join(cwd, "configs", "config.yaml")
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("Could not read config file, proceeding with env variables only", err)
		return
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("Could not parse config file, proceeding with env variables only")
	}
}

func GetString(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		if val, ok := config[key].(string); ok {
			return val
		}
		return defaultValue
	}
	return value
}

func GetInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		if val, ok := config[key].(int); ok {
			return val
		} else if strVal, ok := config[key].(string); ok {
			intValue, err := strconv.Atoi(strVal)
			if err == nil {
				return intValue
			}
		}
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	return intValue
}

func GetBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		if val, ok := config[key].(bool); ok {
			return val
		} else if strVal, ok := config[key].(string); ok {
			boolValue, err := strconv.ParseBool(strVal)
			if err == nil {
				return boolValue
			}
		}
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}
	return boolValue
}
