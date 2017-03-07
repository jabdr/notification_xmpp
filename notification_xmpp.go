package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/mattn/go-xmpp"
	"gopkg.in/yaml.v2"
)

func recvHostBySrv(domain string) (string, error) {
	_, addrs, err := net.LookupSRV("xmpp-client", "tcp", domain)
	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("Could not find any SRV record for %s", domain)
	}

	return fmt.Sprintf("%s:%d", addrs[0].Target, addrs[0].Port), nil
}

type Configuration struct {
	Host              string
	Username          string
	Password          string
	StartTLS          bool
	IgnoreCertificate bool
	Html              bool
}

func NewConfiguration(data []byte) (*Configuration, error) {
	conf := new(Configuration)
	if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, err
	}
	if conf.Username == "" {
		log.Fatalf("You must specify a username in the configuration file!")
	}
	userData := strings.SplitN(conf.Username, "@", 2)
	if len(userData) != 2 {
		log.Fatalf("Invalid username!")
	}
	if conf.Host == "" {
		host, err := recvHostBySrv(userData[1])
		if err != nil {
			conf.Host = userData[1]
		} else {
			conf.Host = host
		}
	}
	return conf, nil
}

func (configuration *Configuration) CreateXMPPClient() (*xmpp.Client, error) {
	options := xmpp.Options{
		Host:     configuration.Host,
		User:     configuration.Username,
		Password: configuration.Password,
		Session:  true,
	}
	if configuration.StartTLS {
		options.NoTLS = true
		options.StartTLS = true
	}
	if configuration.IgnoreCertificate {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return options.NewClient()
}

type Cli struct {
	ConfigurationFile   string
	MessageTemplateFile string
	TargetUsername      string
	Arguments           map[string]string
}

func NewCli() *Cli {
	var cli *Cli

	cli = new(Cli)
	cli.Arguments = make(map[string]string)
	flag.StringVar(&cli.ConfigurationFile, "configuration-file", "", "Path to xmpp configuration file")
	flag.StringVar(&cli.MessageTemplateFile, "template-file", "", "Path to the message template")
	flag.StringVar(&cli.TargetUsername, "target-username", "", "Name of user that should receive the message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: notification_xmpp [options]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()

	if cli.ConfigurationFile == "" {
		log.Fatalf("You must specify a configuration file!")
	}

	if cli.MessageTemplateFile == "" {
		log.Fatalf("You must specify a message template file")
	}

	if cli.TargetUsername == "" {
		log.Fatalf("You mus specify a target username")
	}

	args := flag.Args()
	for _, arg := range args {
		argKv := strings.SplitN(arg, "=", 2)
		if len(argKv) != 2 {
			log.Fatalf("Invalid argument: %s", arg)
		}
		cli.Arguments[argKv[0]] = argKv[1]
	}

	return cli
}

func (cli *Cli) ReadConfiguration() (*Configuration, error) {
	data, err := ioutil.ReadFile(cli.ConfigurationFile)
	if err != nil {
		return nil, err
	}
	conf, err := NewConfiguration(data)
	return conf, err
}

func (cli *Cli) ReadTemplate() (*template.Template, error) {
	return template.ParseFiles(cli.MessageTemplateFile)
}

func sendMessage(cli *Cli, configuration *Configuration) {
	template, err := cli.ReadTemplate()
	if err != nil {
		log.Fatalf("Could not read template: %s", err)
	}

	var templateBuffer bytes.Buffer
	template.Execute(&templateBuffer, cli.Arguments)

	xmppClient, err := configuration.CreateXMPPClient()
	if err != nil {
		log.Fatalf("Could not connect to xmpp server: %s", err)
	}
	defer xmppClient.Close()

	if configuration.Html {
		if _, err = xmppClient.SendHtml(xmpp.Chat{Remote: cli.TargetUsername, Type: "chat", Text: templateBuffer.String()}); err != nil {
			log.Fatalf("Could not send message: %s", err)
		}
	} else {
		if _, err = xmppClient.Send(xmpp.Chat{Remote: cli.TargetUsername, Type: "chat", Text: templateBuffer.String()}); err != nil {
			log.Fatalf("Could not send message: %s", err)
		}
	}
}

func main() {
	cli := NewCli()
	conf, err := cli.ReadConfiguration()
	if err != nil {
		log.Fatalf("Could not read configuration: %s", err)
	}

	sendMessage(cli, conf)
}
