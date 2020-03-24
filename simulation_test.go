package main

import (
	"log"
	"os"
	"reflect"
	"testing"
)

var test_fixtures = "fixtures.yaml"

func fakeLogger() *log.Logger {
	_, logfile, _ := os.Pipe()
	return log.New(logfile, "[Server Test] ", log.LstdFlags)
}

func Test_getConfig(t *testing.T) {
	_, err := getConfig(test_fixtures)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test_Config_GetAllCommandName(t *testing.T) {
	config, _ := getConfig(test_fixtures)
	commands := config.GetAllCommandName()
	if len(commands) != 4 {
		t.Errorf("Incorrect number of commands")
	}
}

func Test_Config_hasCommand(t *testing.T) {
	config, _ := getConfig(test_fixtures)
	if !config.hasCommand("meow") {
		t.Errorf("Not exist command 'meow'")
	}

	if config.hasCommand("nonono") {
		t.Errorf("hasCommand() error")
	}
}

func Test_Config_Verify(t *testing.T) {
	config, _ := getConfig(test_fixtures)
	if config.Verify() != nil {
		t.Errorf("fixtures Verify failed")
	}

	config.Default = append(config.Default, "test_test")

	if config.Verify() == nil {
		t.Errorf("fixtures Verify failed")
	}
}

func Test_Config_DetectMutex(t *testing.T) {
	config, _ := getConfig(test_fixtures)

	for line := range config.DetectMutex("meow") {
		if reflect.DeepEqual(line, []string{"meow", "shine"}) {
			t.Errorf("DetectMutex failed")
		}
	}
}

func Test_GetMethods(t *testing.T) {
	config, _ := getConfig(test_fixtures)

	ch_command := make(chan string)
	ch_send := make(chan []byte, 128)
	ch_recv := make(chan []byte, 128)
	simulation := Simulation{
		Input:   ch_recv,
		Output:  ch_send,
		Command: ch_command,
		config:  config,
		logger:  fakeLogger(),
	}

	go simulation.Run()
	simulation.Command <- "shine"
	if string(<-ch_send) != "No ~ \n" {
		t.Errorf("Output failed")
	}

	simulation.RunDefault()
	if string(<-ch_send) != "Meow ~ Meow ~\n" {
		t.Errorf("Output failed")
	}
}
