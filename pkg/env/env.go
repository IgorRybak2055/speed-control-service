// Package env provide access to env variable
package env

import (
	"os"
)

// GetString return string env variable or passed default value
func GetString(key, defaultValue string) string {
	value, exist := os.LookupEnv(key)
	if !exist {
		return defaultValue
	}

	return value
}
