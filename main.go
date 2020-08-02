package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

var PIDFile = "/tmp/daemonize.pid"

func start(reference string, key string) {
	if _, err := os.Stat(PIDFile); err == nil {
		fmt.Println("Already running or /tmp/daemonize.pid file exist.")
		os.Exit(1)
	}

	cmd := exec.Command(os.Args[0], "mainly")
	cmd.Start()
	fmt.Println("Daemon process ID is : ", cmd.Process.Pid)
	savePID(cmd.Process.Pid)
	fmt.Printf("reference %s\n", reference)
	fmt.Printf("key %s\n", key)
	os.Exit(0)
}

func stop() {
	if _, err := os.Stat(PIDFile); err == nil {
		data, err := ioutil.ReadFile(PIDFile)
		if err != nil {
			fmt.Println("Not running")
			os.Exit(1)
		}
		ProcessID, err := strconv.Atoi(string(data))

		if err != nil {
			fmt.Println("Unable to read and parse process id found in ", PIDFile)
			os.Exit(1)
		}

		process, err := os.FindProcess(ProcessID)

		if err != nil {
			fmt.Printf("Unable to find process ID [%v] with error %v \n", ProcessID, err)
			os.Exit(1)
		}
		// remove PID file
		os.Remove(PIDFile)

		fmt.Printf("Killing process ID [%v] now.\n", ProcessID)
		// kill process and exit immediately
		err = process.Kill()

		if err != nil {
			fmt.Printf("Unable to kill process ID [%v] with error %v \n", ProcessID, err)
			os.Exit(1)
		} else {
			fmt.Printf("Killed process ID [%v]\n", ProcessID)
			os.Exit(0)
		}

	} else {

		fmt.Println("Not running.")
		os.Exit(1)
	}
}
func savePID(pid int) {

	file, err := os.Create(PIDFile)
	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))

	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	file.Sync() // flush to disk

}

func mainly(){
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		signalType := <-ch
		signal.Stop(ch)
		fmt.Println("Exit command received. Exiting...")

		// this is a good place to flush everything to disk
		// before terminating.
		fmt.Println("Received signal type : ", signalType)

		// remove PID file
		os.Remove(PIDFile)

		os.Exit(0)

	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", SayHelloWorld)
	log.Fatalln(http.ListenAndServe(":8080", mux))
}
func SayHelloWorld(w http.ResponseWriter, r *http.Request) {
	html := "Hello World gooooo"

	w.Write([]byte(html))
}
func main() {
	//var language string
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "lang",
				Value: "english",
				Usage: "Language for the greeting",
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "Load configuration from `FILE`",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "start",
				Usage:   "complete a task on the list",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "reference", Aliases: []string{"r"}},
					&cli.StringFlag{Name: "key", Aliases: []string{"k"}},
				},
				Action: func(c *cli.Context) error {

					start(c.String("reference"), c.String("key"))
					return nil
				},
			},
			{
				Name:    "stop",
				Usage:   "add a task to the list",
				Action: func(c *cli.Context) error {
					stop()
					return nil
				},
			},
			{
				Name:    "mainly",
				Usage:   "add a task to the list",
				Action: func(c *cli.Context) error {
					mainly()
					return nil
				},
			},
		},
	}
	// Make arrangement to remove PID file upon receiving the SIGTERM from kill command

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("main")
}
