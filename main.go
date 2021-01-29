package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"text/template"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const messageTemplate = "{{.DateTime}}\nCronic detected failure or error output for the command:\n{{.Cmd}}\n\nRESULT CODE: {{.Code}}\n\nERROR OUTPUT:\n{{.ErrorOut}}\n\nSTANDARD OUTPUT:\n{{.Out}}\n\n{{ if ne .Trace .Out }}\nTRACE-ERROR OUTPUT:\n{{.Trace}}  \n{{ end }}"

type Config struct {
	LogFileName string `envconfig:"LOGFILE_NAME" default:"/var/log/cronic.log"`
}

type dataStruct struct {
	Cmd      string
	Code     int
	ErrorOut string
	Out      string
	Trace    string
	DateTime string
}

func main() {
	s, err := loadConfig()

	args := os.Args
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdin = os.Stdin
	var out bytes.Buffer
	cmd.Stdout = &out
	var outErr bytes.Buffer
	cmd.Stderr = &outErr
	t := template.Must(template.New("messageTemplate").Parse(messageTemplate))

	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			data := dataStruct{
				Cmd:      cmd.String(),
				Code:     waitStatus.ExitStatus(),
				ErrorOut: outErr.String(),
				Out:      out.String(),
				Trace:    outErr.String(),
				DateTime: time.Now().Format(time.RFC3339),
			}
			err := t.Execute(os.Stdout, data)
			if err != nil {
				panic(err)
			}
		}

	}
	data := dataStruct{
		Cmd:      cmd.String(),
		Code:     0,
		ErrorOut: outErr.String(),
		Out:      out.String(),
		Trace:    outErr.String(),
		DateTime: time.Now().Format(time.RFC3339),
	}

	logfile, err := os.OpenFile(s.LogFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
	if err != nil {
		processError(err)
	}

	_, err = fmt.Fprintf(logfile, "\n=================================\n")
	if err != nil {
		processError(err)
	}

	err = t.Execute(logfile, data)
	if err != nil {
		processError(err)
	}
}

func loadConfig() (Config, error) {
	var s Config
	err := envconfig.Process("CRONIC", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	return s, err
}

func processError(err error) {
	fmt.Printf("Failed: %s", err)
}
