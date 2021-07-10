package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	//yaml "gopkg.in/yaml.v2"
	"github.com/manifoldco/promptui"
	yaml "gopkg.in/yaml.v3"
)

func menuSelect(items []string) (string, error) {
	prompt := promptui.Select{
		Label: "Select Command",
		Items: items,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func getConfig(str string) (Config, error) {
	config := Config{}
	configFile, err := ioutil.ReadFile(str)
	yaml.Unmarshal(configFile, &config)
	return config, err
}

func main() {
	help := flag.Bool("h", false, "Show help")
	config_file := flag.String("f", "fixtures.yaml", "the fixtures config")
	config_file = flag.String("c", "fixtures.yaml", "discard")
	isDaemon := flag.Bool("d", false, "No log, No prompt")
	log_path := flag.String("log", "haniel.log", "the running log path")
	socket_server := flag.String("l", "localhost:1234", "As socket Server Address default enable")
	socket_client := flag.String("p", "", "As socket Client Address default disable '-p || -l'")

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	fmt.Println("load fixtures config: ", *config_file)
	config, err := getConfig(*config_file)
	if err != nil {
		log.Println(err)
	}

	if err = config.Verify(); err != nil {
		fmt.Println("=== config error ===")
		fmt.Println(err)
		fmt.Println("=== config error ===")
	}

	logfile, err := os.OpenFile(*log_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panicln(err)
	}
	//logger := log.New(os.Stdout, "[DEV] ", log.LstdFlags)
	logger := log.New(logfile, "[DEV] ", log.LstdFlags)
	if *isDaemon {
		logger = log.New(ioutil.Discard, "", log.LstdFlags)
	}

	ch_command := make(chan string)
	ch_send := make(chan []byte, 128)
	ch_recv := make(chan []byte, 128)
	simulation := Simulation{
		Input:   ch_recv,
		Output:  ch_send,
		Command: ch_command,
		config:  config,
		logger:  logger,
	}

	// Get Menu All Items
	items := config.GetAllCommandName()

	socketServer := &SocketServer{
		logger: logger,
		input:  ch_send,
		output: ch_recv,
	}

	if *socket_client == "" {
		go socketServer.Listen(*socket_server)
	} else {
		go socketServer.Client(*socket_client)
	}

	// Core execute command
	go simulation.Run()

	// deal with socket input
	go simulation.ipc()

	// execute default need run command
	simulation.RunDefault()

	if *isDaemon {
		select {}
	} else {
		for {
			command, err := menuSelect(items)
			if err != nil {
				log.Println(err)
				fmt.Println("Exit")
				break
			}
			fmt.Printf("You choose %q\n", command)
			simulation.Command <- command
		}
	}

}
