package command

import (
	"flag"
	"net/url"

	"code.cloudfoundry.org/cli/plugin"
)

// Logger is used for outputting log-cache results and errors
type Logger interface {
	Fatalf(format string, args ...interface{})
}

func CreateDrain(cli plugin.CliConnection, args []string, log Logger) {
	f := flag.NewFlagSet("cf-syslog-drain", flag.ContinueOnError)
	drainType := f.String("type", "", "")
	err := f.Parse(args)
	if err != nil {
		log.Fatalf("%s", err)
	}

	if len(f.Args()) != 3 {
		log.Fatalf("Invalid arguments, expected 3, got %d.", len(f.Args()))
	}

	appName := f.Args()[0]
	serviceName := f.Args()[1]
	drainURL := f.Args()[2]

	_, err = cli.GetApp(appName)
	if err != nil {
		log.Fatalf("Invalid app name: %s", appName)
	}

	u, err := url.Parse(drainURL)
	if err != nil {
		log.Fatalf("Invalid syslog drain URL: %s", err)
	}

	if *drainType != "" {
		if !validDrainType(*drainType) {
			log.Fatalf("Invalid type: %s", *drainType)
		}

		qValues := u.Query()
		qValues.Set("drain-type", *drainType)
		u.RawQuery = qValues.Encode()
	}

	command := []string{"create-user-provided-service", serviceName, "-l", u.String()}
	_, err = cli.CliCommand(command...)
	if err != nil {
		log.Fatalf("Failed to create service binding: %s", serviceName)
	}

	command = []string{"bind-service", appName, serviceName}
	_, err = cli.CliCommand(command...)
	if err != nil {
		log.Fatalf("Failed to bind %s to %s.", appName, serviceName)
	}
}

func validDrainType(drainType string) bool {
	switch drainType {
	case "logs", "metrics", "all":
		return true
	default:
		return false
	}
}