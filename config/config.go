package config

import (
	"os"
	"strconv"
)

// GetConfig returns the config value or "" if the value does not exist
func GetConfig(name string) string {
	value, exists := LookupConfig(name)
	if !exists {
		return ""
	}
	return value
}

// LookupConfig returns the config value or "" if the value does not exist
func LookupConfig(name string) (string, bool) {
	if configMap == nil {
		value, exists := os.LookupEnv(name)
		if !exists {
			log.LogDebugf("GetConfig name: %v does not exist", name)
			return value, exists
		}
		log.LogDebugf("GetConfig name: %v value: %v", name, value)
		return value, exists
	}
	m, _ := configMap.(map[string]interface{})
	valueString, okString := m[name].(string)
	valueBool, okBool := m[name].(bool)
	if !okString && !okBool {
		log.LogDebugf("GetConfig name: %v does not exist", name)
		return "", false
	}
	value := valueString
	if okBool {
		log.LogDebugf("GetConfig convert bool to string for name: %v", name)
		value = strconv.FormatBool(valueBool)
	}

	log.LogDebugf("GetConfig name: %v value: %v", name, value)
	return value, true
}

// // LookupConfig1 returns the config value or false if the value does not exist
// func LookupConfig1(name string) (string, bool) {
// 	if configMap == nil {
// 		return os.LookupEnv(name)
// 	}

// 	m, _ := configMap.(map[string]interface{})
// 	_, okString := m[name].(string)
// 	_, okBool := m[name].(bool)
// 	if !okString && !okBool {
// 		log.LogDebugf("GetConfig could not find name: %v", name)
// 		return "", false
// 	}
// 	return GetConfig(name), true
// }
