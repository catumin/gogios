package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

var (
	inittime   = time.Now()
	password   = flag.String("password", "password123", "Password to use. Will also be used as the passphrase for encrypted keys")
	authkey    = flag.String("privatekey", "/dev/null", "Private key to use")
	authmethod = flag.String("authmeth", "password", "Authentication method to use. password, eauthkey, or uauthkey")
	host       = flag.String("host", "123.123.123.123", "IP address or FQDN to connect to")
	port       = flag.Int("port", 22, "Port of server")
	user       = flag.String("user", "root", "User to login as")
	attempts   = flag.Int("attempts", 3, "Amount of times to attempt login")
	timer      = flag.Duration("timer", 300*time.Millisecond, "Timeout between attempts")
)

type resp struct {
	Error error
	mu    sync.Mutex
}

// sshKey - Attempt to authenticate with an SSH server using a keyfile to confirm it is accessible
func sshKey(key ssh.AuthMethod) *resp {
	exitcode := &resp{}

	config := &ssh.ClientConfig{
		User:            *user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{key},
		Timeout:         *timer,
	}

	_, err := ssh.Dial("tcp", *host+":"+strconv.Itoa(*port), config)
	if err != nil {
		fmt.Println("\nFailed connection")
	} else {
		end := time.Now()
		d := end.Sub(inittime)
		duration := d.Seconds()
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Fprintf(color.Output, "\n%s", color.GreenString("Successful login"))
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Printf("\nCompleted in %v seconds\n", strconv.FormatFloat(duration, 'g', -1, 64))
	}

	exitcode.Error = err
	return exitcode
}

// sshPassword - Attempt to authenticate with an SSH server using a password to confirm it is accessible
func sshPassword(password string) *resp {
	exitcode := &resp{}

	config := &ssh.ClientConfig{
		User:            *user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		Timeout:         *timer,
	}

	_, err := ssh.Dial("tcp", *host+":"+strconv.Itoa(*port), config)
	if err != nil {
		fmt.Println("\nFailed connection")
	} else {
		end := time.Now()
		d := end.Sub(inittime)
		duration := d.Seconds()
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Fprintf(color.Output, "\n%s", color.GreenString("Successful login"))
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Printf("\nCompleted in %v seconds\n", strconv.FormatFloat(duration, 'g', -1, 64))
	}

	exitcode.Error = err
	return exitcode
}

func main() {
	flag.Parse()

	switch authmethod := *authmethod; authmethod {
	case "password":
		for attempt := *attempts; attempt != 0; attempt-- {
			go func() {
				resp := sshPassword(*password)
				resp.mu.Lock()
				if resp.Error == nil {
					resp.mu.Unlock()
					os.Exit(0)
				}
			}()

			fmt.Println("Attempts left: ", attempt)
			time.Sleep(*timer)
		}
	case "uauthkey":
		for attempt := *attempts; attempt != 0; attempt-- {
			go func() {
				resp := sshKey(KeyFile(*authkey))
				resp.mu.Lock()
				if resp.Error == nil {
					resp.mu.Unlock()
					os.Exit(0)
				}
			}()

			fmt.Println("Attempts left: ", attempt)
			time.Sleep(*timer)
		}
	case "eauthkey":
		for attempt := *attempts; attempt != 0; attempt-- {
			go func() {
				resp := sshKey(EncryptedKeyFile(*authkey, *password))
				resp.mu.Lock()
				if resp.Error == nil {
					resp.mu.Unlock()
					os.Exit(0)
				}
			}()

			fmt.Println("Attempts left: ", attempt)
			time.Sleep(*timer)
		}
	default:
		fmt.Println("Invalid auth method ", authmethod)
		os.Exit(1)
	}
}
