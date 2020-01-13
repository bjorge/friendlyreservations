package config

import (
	"testing"
)

func TestConfigFile(t *testing.T) {
	t.Log("config is:")
	t.Log(configMap)
	t.Log("end of config")
	value1 := GetConfig("SETTING1")
	if value1 != "value1" {
		t.Errorf("value1 not found")
	}
	value2 := GetConfig("SETTING2")
	if value2 != "value2" {
		t.Errorf("value2 not found")
	}
	value2, exists := LookupConfig("SETTING2")
	if !exists {
		t.Errorf("value2 not exists but should exist")
	}
	if value2 != "value2" {
		t.Errorf("value2 not found")
	}
	_, exists = LookupConfig("SETTINGS3")
	if exists {
		t.Errorf("value3 should not exist")
	}

	valueBool := GetConfig("SETTINGBOOL")
	if valueBool != "true" {
		t.Errorf("value bool not true, it is %v", valueBool)
	}
}
