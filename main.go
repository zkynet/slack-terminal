package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nlopes/slack"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error loading .env file")
	}

	api := slack.New(os.Getenv("SLACK_API_KEY"))
	//logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	//slack.SetLogger(logger)
	//api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			botTagString := fmt.Sprintf("<@%s>", rtm.GetInfo().User.ID)
			if !strings.Contains(ev.Msg.Text, botTagString) {
				continue
			}
			message := strings.Replace(ev.Msg.Text, botTagString, "", -1)
			messageParts := strings.Split(message, " ")
			if messageParts[1] == "remote" {
				go remoteExecution(messageParts, rtm, ev)
			} else if messageParts[1] == "local" {
				go localExecution(messageParts, rtm, ev)
			}

		default:
			break
		}
	}

}

func localExecution(messageParts []string, rtm *slack.RTM, ev interface{}) {

	var cmd *exec.Cmd
	cmd = exec.Command(messageParts[2], strings.Join(messageParts[3:], " "))
	fmt.Println("Registering local command:" + strings.Join(messageParts[2:], " "))
	copyAndCaptureOutput(cmd, rtm, ev)
	cmd.Start()

	err := cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

}

func remoteExecution(messageParts []string, rtm *slack.RTM, ev interface{}) {

	var cmd *exec.Cmd
	cmd = exec.Command(os.Getenv("SSH_COMMAND_PATH"),
		os.Getenv("SSH_COMMAND_ARGS"),
		"ssh -p "+os.Getenv("SSH_PORT")+" -i "+os.Getenv("SSH_KEY")+" "+os.Getenv("SSH_USER")+"@"+messageParts[2]+" 'sudo -u root "+strings.Join(messageParts[3:], " ")+";exit'")
	fmt.Println("Registering remote command:" + messageParts[2] + " : " + strings.Join(messageParts[3:], " "))
	copyAndCaptureOutput(cmd, rtm, ev)
	cmd.Start()

	err := cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

}

func copyAndCaptureOutput(cmd *exec.Cmd, rtm *slack.RTM, ev interface{}) {

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	go func() {
		_, _ = copyAndCapture(os.Stdout, stdoutIn, rtm, ev)
	}()

	go func() {
		_, _ = copyAndCapture(os.Stderr, stderrIn, rtm, ev)
	}()
}

func copyAndCapture(w io.Writer, r io.Reader, rtm *slack.RTM, ev interface{}) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			rtm.SendMessage(rtm.NewOutgoingMessage(string(d), ev.(*slack.MessageEvent).Channel))
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
	// never reached
	panic(true)
	return nil, nil
}
