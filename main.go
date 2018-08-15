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
			//fmt.Println(msg.Data)
			messageParts := strings.Split(message, " ")
			fmt.Println()

			var cmd *exec.Cmd
			cmd = exec.Command(os.Getenv("SSH_COMMAND_PATH"),
				os.Getenv("SSH_COMMAND_ARGS"),
				"ssh -p "+os.Getenv("SSH_PORT")+" -i "+os.Getenv("SSH_KEY")+" "+os.Getenv("SSH_USER")+"@"+messageParts[1]+" 'sudo -u root "+strings.Join(messageParts[2:], " ")+";exit'")
			fmt.Println("Registering command:" + messageParts[1] + " : " + strings.Join(messageParts[2:], " "))
			var stdout, stderr []byte
			var errStdout, errStderr error
			stdoutIn, _ := cmd.StdoutPipe()
			stderrIn, _ := cmd.StderrPipe()
			cmd.Start()

			go func() {
				stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn, rtm, ev)
			}()

			go func() {
				stderr, errStderr = copyAndCapture(os.Stderr, stderrIn, rtm, ev)
			}()

			err := cmd.Wait()
			if err != nil {
				log.Fatalf("cmd.Run() failed with %s\n", err)
			}
			if errStdout != nil || errStderr != nil {
				log.Fatalf("failed to capture stdout or stderr\n")
			}

		default:
			break
		}
	}

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
			//_, err := w.Write(d)
			//if err != nil {
			//	return out, err
			//}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
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
