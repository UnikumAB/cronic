package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

//go:embed message.tmpl
var messageTemplate string

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
	t := template.Must(template.New("messageTemplate").Parse(messageTemplate))

	s, err := loadConfig()
	if err != nil {
		processError(err)
	}
	versionFlag := flag.Bool("version", false, "Displays the version of cronic and exit")
	logfileFlag := flag.String("logfile", s.LogFileName, "Logfile name. Defaults to value of the CRONIC_LOGFILE_NAME env variable.")
	flag.Parse()
	if *versionFlag {
		fmt.Printf("Cronic version %s, commit=%q, build on %s by %s\n", version, commit, date, builtBy)
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	var out bytes.Buffer
	cmd.Stdout = &out
	var outStdErr bytes.Buffer
	cmd.Stderr = &outStdErr

	err = cmd.Run()
	outErrString := outStdErr.String()
	outTrace, outErr := filterErrorOutput(outErrString)

	stdOutString := out.String()
	data := dataStruct{
		Cmd:      cmd.String(),
		Code:     0,
		ErrorOut: outErr,
		Out:      stdOutString,
		Trace:    outTrace,
		DateTime: time.Now().Format(time.RFC3339),
	}
	if err != nil {
		log.Printf("Got error: %s", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Printf("Got exitError")
			data.Code = exitError.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			data.ErrorOut = err.Error()
			data.Code = -1
		}
	}

	if data.Code != 0 {
		err := t.Execute(os.Stdout, data)
		if err != nil {
			processError(err)
		}
	}
	logfile, err := os.OpenFile(*logfileFlag, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
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

// filterErrorOutput takes the output and looks for line on Trace level. Then it outputs those as seperate strings.
func filterErrorOutput(outStdErr string) (string, string) {
	var outTrace bytes.Buffer
	var outErr bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(outStdErr))
	ps4 := getEnvOrDefault("PS4", "+ ")
	tracePattern := fmt.Sprintf("^%s+%s", regexp.QuoteMeta(string(ps4[0])), regexp.QuoteMeta(string(ps4[1])))
	pattern, err := regexp.Compile(tracePattern)
	if err != nil {
		processError(err)
	}
	for scanner.Scan() {
		text := scanner.Text()
		if pattern.MatchString(text) {
			outTrace.Write([]byte(text))
			outTrace.Write([]byte("\n"))
			fmt.Print(".")
		} else {
			fmt.Print("+")
			outErr.Write([]byte(text))
			outErr.Write([]byte("\n"))
		}
	}
	return outTrace.String(), outErr.String()
}

func getEnvOrDefault(envKey string, defaultValue string) string {
	ps4 := os.Getenv(envKey)
	if ps4 == "" {
		ps4 = defaultValue
	}
	return ps4
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
	log.Fatalf("failed: %s", err)
}
