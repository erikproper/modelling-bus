/*
 *
 * Package: mbconnect
 * Layer:   generic
 * Module:  config_reader
 *
 * ..... ... .. .
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: XX.11.2025
 *
 */

package mbconnect

import (
	"gopkg.in/ini.v1"
)

type (
	TConfigData struct {
		errorReporter TErrorReporter
		iniFile       *ini.File
	}

	TConfigValue struct {
		errorReporter TErrorReporter
		configKey     *ini.Key
	}
)

func LoadConfig(filePath string, errorReporter TErrorReporter) (*TConfigData, bool) {
	var (
		err        error
		configData TConfigData
	)

	configData.errorReporter = errorReporter
	configData.iniFile, err = ini.Load(filePath)

	if err != nil {
		configData.errorReporter("Failed to read config file:", err)
		configData.iniFile = nil
		return &configData, false
	}

	return &configData, true
}

func (c *TConfigData) GetValue(section, key string) *TConfigValue {
	var configValue TConfigValue

	configValue.configKey = c.iniFile.Section(section).Key(key)

	return &configValue
}

func (v *TConfigValue) StringWithDefault(defaultString string) string {
	s := v.configKey.String()
	if s == "" {
		return defaultString
	} else {
		return s
	}
}

func (v *TConfigValue) String() string {
	return v.StringWithDefault("")
}

func (v *TConfigValue) IntWithDefault(defaultInt int) int {
	i, err := v.configKey.Int()
	if err == nil {
		return i
	} else {
		return defaultInt
	}
}

func (v *TConfigValue) Int() int {
	return v.IntWithDefault(0)
}

func IntWithDefault(key *ini.Key, defaultInt int) int {
	keyInt, err := key.Int()
	if err == nil {
		return keyInt
	} else {
		return defaultInt
	}
}
