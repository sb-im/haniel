package main

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"
)

const (
	CMD     = "cmd"
	PUTS    = "puts"
	PUT_HEX = "put_hex"
	SLEEP   = "sleep"
)

type Config struct {
	Commands map[string][]map[string]string
	Input    []map[string]string
	Default  []string
	Mutex    [][]string
}

func (this *Config) GetAllCommandName() []string {
	var commands []string
	for command, _ := range this.Commands {
		commands = append(commands, command)
	}
	return commands
}

func (this *Config) hasCommand(cmd string) bool {
	for command, _ := range this.Commands {
		if command == cmd {
			return true
		}
	}
	return false
}

func (this *Config) Verify() error {
	var store []string
	// input
	for _, item := range this.Input {
		for command, _ := range item {
			if !this.hasCommand(command) {
				store = append(store, "Input: No Command "+command)
			}
		}
	}

	// default
	for _, command := range this.Default {
		if !this.hasCommand(command) {
			store = append(store, "Default: No Command "+command)
		}
	}

	// mutex
	for _, item := range this.Mutex {
		for _, command := range item {
			if !this.hasCommand(command) {
				store = append(store, "Mutex: No Command "+command)
			}
		}
	}

	if len(store) == 0 {
		return nil
	}
	return errors.New(strings.Join(store, "\n"))
}

func (this *Config) DetectMutex(command string) [][]string {
	var results [][]string
	for _, m := range this.Mutex {
		for _, i := range m {
			if command == i {
				results = append(results, m)
			}
		}
	}
	return results
}

type Simulation struct {
	running map[string]context.CancelFunc
	config  Config
	logger  *log.Logger
	Input   chan []byte
	Output  chan []byte
	Command chan string
}

func (this *Simulation) RunDefault() {
	for _, command := range this.config.Default {
		this.Command <- command
	}
}

func (this *Simulation) ipc() {
	for raw := range this.Input {
		match := false

		for _, item := range this.config.Input {
			for k, v := range item {
				if v == string(raw) {
					match = true
					this.Command <- k
				}
			}
		}

		if !match {
			this.logger.Printf("Ignore Input: %v\n", string(raw))
		}

	}
}

func (this *Simulation) Run() {
	this.running = make(map[string]context.CancelFunc)
	for command := range this.Command {

		for _, line := range this.config.DetectMutex(command) {
			for _, i := range line {

				// Cancel All Mutex
				if this.running[i] != nil {
					this.running[i]()
					delete(this.running, i)
				}

			}
		}

		// Cancel duplicate
		if this.running[command] != nil {
			this.running[command]()
		}

		// New Cancel() save to this.running
		ctx, cancel := context.WithCancel(context.Background())
		this.running[command] = cancel

		go this.doCommand(command, ctx)
	}
}

func (this *Simulation) doCommand(command string, ctx context.Context) {
	this.logger.Printf("=== RUN %v RUN === \n", command)

	defer func() {
		this.logger.Printf("=== END %v END === \n", command)
	}()

	for _, value := range this.config.Commands[command] {
		select {
		case <-ctx.Done():
			this.logger.Printf("Executed cancel() \n")
			return
		default:
			//this.logger.Println("Exec: ", value)

			switch {
			case value[PUTS] != "":
				this.logger.Printf("PUTS: %v\n", value[PUTS])

				this.Output <- []byte(value[PUTS] + "\n")
			case value[PUT_HEX] != "":
				this.logger.Printf("PUT_HEX: %v\n", value[PUT_HEX])

				str, err := hex.DecodeString(value[PUT_HEX])
				if err != nil {
					this.logger.Panicln(err)
				}
				this.Output <- []byte(str)

			case value[SLEEP] != "":
				this.logger.Printf("SLEEP: %v\n", value[SLEEP])

				if second, err := time.ParseDuration(value[SLEEP]); err != nil {
					this.logger.Println("ERROR: ", err)
				} else {
					time.Sleep(second)
				}

			case value[CMD] != "":
				this.logger.Printf("CMD: %v\n", value[CMD])

				//go this.doCommand(value[CMD], ctx)
				this.Command <- value[CMD]
			default:
				this.logger.Println("Unknow: ", value)
			}

		}
	}
}
