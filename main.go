package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"kcapture/models"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

//LogMessage prints in JSON format
func LogMessage(logLevel, message string) {

	logStruct := models.LogFormat{Loglevel: logLevel, Message: message}
	logStr, _ := json.Marshal(logStruct)
	log.Println(string(logStr))

}

//pingTest will run test pings to a destination and report on packet loss
func pingTest(destination string, count string) error {
	log.Println("starting ping...")
	cmd := exec.Command("ping", destination, "-c "+count)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	//cmd.Run()
	output, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	//fmt.Printf("output:\n%s\nerr:\n%s\n", output, errStr)

	if err != nil {
		LogMessage("ERROR", output+"\n"+errStr)
		return err
	} else if strings.Contains(string(output), "timeout") {
		LogMessage("ERROR", output+"\n"+errStr)
		err := errors.New(output + "\n" + errStr)
		return err
	} else if strings.Contains(string(output), "Destination Host Unreachable") {
		LogMessage("ERROR", output+"\n"+errStr)
		err := errors.New(output + "\n" + errStr)
		return err
	}

	LogMessage("INFO", "no packet loss")
	LogMessage("INFO", string(output))

	return nil
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

func main() {

	// destination := flag.String("destination", "google.com", "URL and google.com by default")
	// count := flag.Int("count", 10, "number of concurrent requestCount")
	// //default is zero and will use an infinite loop
	// loop := flag.Int("loop", 0, "number of iterations")

	// flag.Parse()

	// strCount := strconv.Itoa(*count)

	//Initialize variables
	destination := GetEnvVar("PING_DESTINATION", "google.com")
	count := GetEnvVar("PING_COUNT", "10")
	loopStr := GetEnvVar("LOOP", "0")
	loop, err := strconv.Atoi(loopStr)
	if err != nil {
		log.Fatal("Need an integer for the LOOP variable.")
	}

	fmt.Printf("\ncount:%s", count)
	fmt.Printf("\ndestinationt:%s\n", destination)

	//loop x number of times otherwise continue
	if loop != 0 {
		for i := 0; i < loop; i++ {

			err := pingTest(destination, count)
			if err != nil {

				LogMessage("ERROR", err.Error())
				sendEmail("ERROR sending pings")
			}
		}
	} else {
		for {
			err := pingTest(destination, count)
			if err != nil {
				LogMessage("ERROR", err.Error())
				sendEmail("ERROR sending pings")
			}
			time.Sleep(5 * time.Second)
		}
	}
}
