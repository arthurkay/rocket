package client

import (
	"flag"
	"fmt"
	"os"
	"rocket/version"
)

const usage1 string = `Usage: %s [OPTIONS] <local port or address>
Options:
`

const usage2 string = `
Examples:
	rocket 80
	rocket -subdomain=example 8080
	rocket -proto=tcp 22
	rocket -hostname="example.com" -httpauth="user:password" 10.0.0.1


Advanced usage: rocket [OPTIONS] <command> [command args] [...]
Commands:
	rocket start [tunnel] [...]    Start tunnels by name from config file
	rocket start-all               Start all tunnels defined in config file
	rocket list                    List tunnel names from config file
	rocket help                    Print help
	rocket version                 Print rocket version

Examples:
	rocket start www api blog pubsub
	rocket -log=stdout -config=rocket.yml start ssh
	rocket start-all
	rocket version

`

type Options struct {
	config        string
	logto         string
	loglevel      string
	authtoken     string
	httpauth      string
	hostname      string
	protocol      string
	subdomain     string
	command       string
	inspectaddr   string
	serveraddr    string
	inspectpublic bool
	tls           bool
	tlsClientCrt  string
	tlsClientKey  string
	args          []string
}

func ParseArgs() (opts *Options, err error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage1, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, usage2)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	config := flag.String(
		"config",
		fmt.Sprintf("%s/.rocket", home),
		"Path to rocket configuration file. (default: $HOME/.rocket)")

	logto := flag.String(
		"log",
		"stdout",
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	loglevel := flag.String(
		"log-level",
		"DEBUG",
		"The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")

	authtoken := flag.String(
		"authtoken",
		"",
		"Authentication token for identifying an rocket account")

	httpauth := flag.String(
		"httpauth",
		"",
		"username:password HTTP basic auth creds protecting the public tunnel endpoint")

	subdomain := flag.String(
		"subdomain",
		"",
		"Request a custom subdomain from the rocket server.")

	hostname := flag.String(
		"hostname",
		"",
		"Request a custom hostname from the rocket server. (HTTP only) (requires CNAME of your DNS)")

	protocol := flag.String(
		"proto",
		"http+https",
		"The protocol of the traffic over the tunnel (http+https|https|tcp)")

	tls := flag.Bool(
		"tls",
		false,
		"Use dial for tls port")

	tlsClientCrt := flag.String(
		"tlsClientCrt",
		"",
		"Path to a TLS Client CRT file if server requires")

	tlsClientKey := flag.String(
		"tlsClientKey",
		"",
		"Path to a TLS Client Key file if server requires")

	serveraddr := flag.String(
		"serveraddr",
		"",
		"The addr for server")

	inspectaddr := flag.String(
		"inspectaddr",
		defaultInspectAddr,
		"The addr for inspect requests")

	inspectpublic := flag.Bool(
		"inspectpublic",
		false,
		"Should export inspector to public access")

	flag.Parse()

	opts = &Options{
		config:        *config,
		logto:         *logto,
		loglevel:      *loglevel,
		httpauth:      *httpauth,
		subdomain:     *subdomain,
		protocol:      *protocol,
		authtoken:     *authtoken,
		hostname:      *hostname,
		serveraddr:    *serveraddr,
		inspectaddr:   *inspectaddr,
		inspectpublic: *inspectpublic,
		tls:           *tls,
		tlsClientCrt:  *tlsClientCrt,
		tlsClientKey:  *tlsClientKey,
		command:       flag.Arg(0),
	}

	switch opts.command {
	case "list":
		opts.args = flag.Args()[1:]
	case "start":
		opts.args = flag.Args()[1:]
	case "start-all":
		opts.args = flag.Args()[1:]
	case "version":
		fmt.Println(version.MajorMinor())
		os.Exit(0)
	case "help":
		flag.Usage()
		os.Exit(0)
	case "":
		err = fmt.Errorf("error: Specify a local port to tunnel to, or " +
			"a rocket command.\n\nExample: To expose port 80, run " +
			"'rocket 80'")
		return

	default:
		if len(flag.Args()) > 1 {
			err = fmt.Errorf("you may only specify one port to tunnel to on the command line, got %d: %v",
				len(flag.Args()),
				flag.Args())
			return
		}

		opts.command = "default"
		opts.args = flag.Args()
	}

	return
}
