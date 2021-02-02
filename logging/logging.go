package logging

import (
	"io"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/logfmt"
	"github.com/apex/log/handlers/text"
	isatty "github.com/mattn/go-isatty"
	"github.com/pkg/errors"
)

func SetupLog(format, output string, debug bool) error {

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	var out io.Writer
	switch output {
	case "":
		out = os.Stderr
	case "-":
		out = os.Stdout
	default:
		var err error
		out, err = os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return errors.Wrapf(err, "opening log file '%s' for appending", output)
		}
	}

	if format == "" {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			format = "dev"
		} else {
			format = "json"
		}
	}

	var handler log.Handler
	switch format {
	case "json":
		handler = json.New(out)
	case "log":
		handler = logfmt.New(out)
	case "dev":
		handler = text.New(out)
	case "cli":
		handler = cli.New(out)
	default:
		return errors.Errorf("unknown log format '%s'", format)
	}
	log.SetHandler(handler)

	return nil

}
