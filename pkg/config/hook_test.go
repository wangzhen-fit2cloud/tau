package config

import (
	"fmt"
	"testing"

	"github.com/avinor/tau/pkg/helpers/strings"
	"github.com/stretchr/testify/assert"
)

const (
	hookTest1 = `
		hook "name" {
			command = "command"
			trigger_on = "prepare"
		}
	`

	hookTest2 = `
		hook "name" {
			command = "overwrite"
			args = ["arg1", "arg2"]
		}
	`

	hookTest3 = `
		hook "name" {
			args = ["arg3"]
		}
	`

	hookTest4 = `
		hook "name" {
			command = "command"
			trigger_on = "invalid"
		}
	`

	hookTest5 = `
		hook "name" {
			command = ""
			trigger_on = "prepare"
		}
	`

	hookTest6 = `
		hook "name" {
			command = "test"
			script = "path"
			trigger_on = "prepare"
		}
	`

	hookTest7 = `
		hook "name" {
			script = "path"
			trigger_on = "prepare:init"
		}
	`
)

var (
	hookFile1, _ = NewFile("/hook1", []byte(hookTest1))
	hookFile2, _ = NewFile("/hook2", []byte(hookTest2))
	hookFile3, _ = NewFile("/hook3", []byte(hookTest3))
	hookFile4, _ = NewFile("/hook4", []byte(hookTest4))
	hookFile5, _ = NewFile("/hook5", []byte(hookTest5))
	hookFile6, _ = NewFile("/hook6", []byte(hookTest6))
	hookFile7, _ = NewFile("/hook7", []byte(hookTest7))
)

func TestHookMerge(t *testing.T) {
	tests := []struct {
		Files    []*File
		Expected []*Hook
		Error    error
	}{
		{
			[]*File{hookFile1},
			[]*Hook{
				{
					Type:      "name",
					Command:   strings.ToPointer("command"),
					TriggerOn: strings.ToPointer("prepare"),
				},
			},
			nil,
		},
		{
			[]*File{hookFile1, hookFile2},
			[]*Hook{
				{
					Type:      "name",
					Command:   strings.ToPointer("overwrite"),
					TriggerOn: strings.ToPointer("prepare"),
					Arguments: &[]string{"arg1", "arg2"},
				},
			},
			nil,
		},
		{
			[]*File{hookFile1, hookFile2, hookFile3},
			[]*Hook{
				{
					Type:      "name",
					Command:   strings.ToPointer("overwrite"),
					TriggerOn: strings.ToPointer("prepare"),
					Arguments: &[]string{"arg1", "arg2", "arg3"},
				},
			},
			nil,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			config := &Config{}
			err := mergeHooks(config, getConfigFromFiles(t, test.Files))

			if test.Error != nil {
				assert.EqualError(t, err, test.Error.Error())
				return
			}

			actual := map[string]string{}
			for _, dep := range config.Dependencies {
				actual[dep.Name] = dep.Source
			}

			assert.ElementsMatch(t, test.Expected, config.Hooks)
		})
	}
}

func TestHookValidation(t *testing.T) {
	tests := []struct {
		Files    []*File
		Expected map[string]ValidationResult
	}{
		{
			[]*File{hookFile1},
			map[string]ValidationResult{
				"name": {Result: true, Error: nil},
			},
		},
		{
			[]*File{hookFile7},
			map[string]ValidationResult{
				"name": {Result: true, Error: nil},
			},
		},
		{
			[]*File{hookFile4},
			map[string]ValidationResult{
				"name": {Result: false, Error: triggerOnValueIncorrect},
			},
		},
		{
			[]*File{hookFile5},
			map[string]ValidationResult{
				"name": {Result: false, Error: scriptOrCommandIsRequired},
			},
		},
		{
			[]*File{hookFile6},
			map[string]ValidationResult{
				"name": {Result: false, Error: scriptAndCommandBothDefined},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			config := &Config{}
			err := mergeHooks(config, getConfigFromFiles(t, test.Files))
			assert.NoError(t, err)

			for _, hook := range config.Hooks {
				result, err := hook.Validate()
				expect := test.Expected[hook.Type]

				assert.Equal(t, expect.Result, result)
				if expect.Error != nil {
					assert.EqualError(t, err, expect.Error.Error())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}
