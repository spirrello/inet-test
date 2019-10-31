package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
)

// smtpServer data to smtp server.
type smtpServer struct {
	host string
	port string
}

// Address URI to smtp server.
func (s *smtpServer) Address() string {
	return s.host + ":" + s.port
}

//GetEnvVar is a helper function for gathering env variables
func GetEnvVar(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

type logFormat struct {
	Loglevel string `json:"level"`
	Caller   string `json:"caller"`
	Message  string `json:"message"`
}

//LogMessage prints in JSON format
func LogMessage(logLevel, message string) {

	//identify calling function
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)

	logStruct := logFormat{Loglevel: logLevel, Caller: details.Name(), Message: message}
	logStr, _ := json.Marshal(logStruct)
	log.Println(string(logStr))

}

//runCommand runs the ping test and greps appropriately for packet loss
func runCommand(commandApp string, args []string) (string, string, error) {
	log.Println(args[0])
	log.Println(args[1])
	cmd := exec.Command(commandApp, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		LogMessage("ERROR", "runCommand:"+string(stderr.Bytes()))
		return "", "", err
	}

	return string(stdout.Bytes()), string(stderr.Bytes()), nil

}

//pingTest will run test pings to a destination and report on packet loss
func pingTest(destination string, count string) (string, error) {
	log.Println("starting ping...")

	var grepStr string
	// var err error
	if runtime.GOOS == "darwin" {
		grepStr = "7"
	} else {
		grepStr = "6"
	}

	output, err := sh.Command("ping", destination, "-c", count).Command("grep", "loss").Command("awk", "{print $"+grepStr+"}").Output()

	//log errors and return
	if err != nil {
		LogMessage("ERROR", err.Error())
		return string(output), err
	}

	outputStr := strings.TrimSpace(string(output))

	return outputStr, nil
}

//sendeEmail will send email alerts
func sendEmail(message string) {

	//set variables from ENV
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	senderEmail := os.Getenv("SENDER_EMAIL")
	destEmail := []string{os.Getenv("DEST_EMAIL")}
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	smtpServer := smtpServer{host: smtpHost, port: smtpPort}

	emailMessage := []byte(message)

	auth := smtp.PlainAuth("", senderEmail, emailPassword, smtpServer.host)

	err := smtp.SendMail(smtpServer.Address(), auth, senderEmail, destEmail, emailMessage)
	if err != nil {
		// fmt.Println(err)
		LogMessage("ERROR", err.Error())
		return
	}
	fmt.Println("Email Sent!")

}

//parsePingTimeResult checks results for errors
func parsePingTestResult(pingResult string, err error) {

	if err != nil {
		LogMessage("ERROR", "pingResult:"+pingResult+" err:"+err.Error())
		sendEmail("pingResult:" + pingResult + " err:" + err.Error())
		return
	} else if pingResult != "0%" && pingResult != "0.0%" {
		//convert to int and check if packet loss is higher than 10% send email alerts
		LogMessage("ERROR", "packet loss: "+pingResult)
		pingResult = strings.Trim(pingResult, "%")
		pingResultInt, intErr := strconv.ParseFloat(pingResult, 64)

		if intErr != nil {
			LogMessage("ERROR", "coudln't convert pingResult: "+intErr.Error())
			return
		} else if pingResultInt >= 10 {
			sendEmail("packet loss: " + pingResult)
			return
		}
	}

	// LogMessage("INFO", "packet loss: "+pingResult)
	return

}

func main() {

	//Initialize variables
	destination := GetEnvVar("PING_DESTINATION", "google.com")
	count := GetEnvVar("PING_COUNT", "10")
	loopStr := GetEnvVar("LOOP", "0")
	loop, err := strconv.Atoi(loopStr)
	if err != nil {
		log.Fatal("Need an integer for the LOOP variable.")
	}

	fmt.Printf("\ncount:%s", count)
	fmt.Printf("\ndestination:%s\n", destination)

	//loop x number of times otherwise continue
	if loop != 0 {
		for i := 0; i < loop; i++ {
			//invoke pingTest and then parse results
			pingResult, err := pingTest(destination, count)
			parsePingTestResult(pingResult, err)
		}
	} else {
		for {
			//invoke pingTest and then parse results
			pingResult, err := pingTest(destination, count)
			parsePingTestResult(pingResult, err)
			time.Sleep(5 * time.Second)
		}
	}
}
