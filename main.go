package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kcapture/models"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"runtime"
	"strconv"
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

	var output, errStr, grepStr string
	var err error
	if runtime.GOOS == "darwin" {
		grepStr = " | grep loss | awk '{print $7}'"
	} else {
		grepStr = " | grep loss | awk '{print $6}'"
	}
	fmt.Println(grepStr)
	args := []string{destination, "-c " + count}
	output, errStr, err = runCommand("ping", args)

	//runCommand(destination, count, grepStr)
	// cmd := exec.Command("ping", destination, "-c "+count+"| grep loss | awk '{print $7}'")
	// cmd := exec.Command("ping", destination, "-c "+count)
	// var stdout, stderr bytes.Buffer
	// cmd.Stdout = &stdout
	// cmd.Stderr = &stderr
	// err := cmd.Run()
	// output, errStr := string(stdout.Bytes()), string(stderr.Bytes())

	//log errors and return
	if err != nil {
		LogMessage("ERROR", "pingTest:"+errStr)
		return "", err
	}
	// } else if strings.Contains(string(output), "timeout") {
	// 	LogMessage("ERROR", output+"\n"+errStr)
	// 	err := errors.New(output + "\n" + errStr)
	// 	return err
	// } else if strings.Contains(string(output), "Destination Host Unreachable") {
	// 	LogMessage("ERROR", output+"\n"+errStr)
	// 	err := errors.New(output + "\n" + errStr)
	// 	return err
	// }

	LogMessage("INFO", "packet loss:"+string(output))

	return output, nil
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

			pingResult, err := pingTest(destination, count)
			if err != nil {
				LogMessage("ERROR", err.Error())
				//sendEmail("ERROR sending pings")
			} else if pingResult != "0%" {
				LogMessage("ERROR", err.Error())
				//sendEmail("packet loss: " + pingResult)
			}
		}
	} else {
		for {
			pingResult, err := pingTest(destination, count)
			if err != nil {
				LogMessage("ERROR", err.Error())
				//sendEmail("ERROR sending pings")
			} else if pingResult != "0%" {
				//LogMessage("ERROR", err.Error())
				//sendEmail("packet loss: " + pingResult)
			}
			time.Sleep(5 * time.Second)
		}
	}
}
