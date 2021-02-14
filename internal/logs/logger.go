package logs

import (
	"os"

	"github.com/mtekmir/warehouse-service/internal/errors"
	"github.com/sirupsen/logrus"
)

// NewLogger creates a new logger instance.
func NewLogger(env, logfile *string) (*logrus.Logger, error) {
	log := logrus.New()

	if *env == "local" {
		log.Out = os.Stdout
		return log, nil
	}

	if logfile != nil {
		file, err := os.OpenFile(*logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, errors.E("Unable to open logfile", err)
		}
		log.Out = file
	}

	log.SetFormatter(&formatter{
		fields: logrus.Fields{
			"service": "warehouse-service",
			"env":     env,
		},
		lf: &logrus.JSONFormatter{},
	})

	return log, nil
}

// formatter adds default fields to each log entry.
type formatter struct {
	fields logrus.Fields
	lf     logrus.Formatter
}

// Format satisfies the logrus.Formatter interface.
func (f *formatter) Format(e *logrus.Entry) ([]byte, error) {
	for k, v := range f.fields {
		e.Data[k] = v
	}
	return f.lf.Format(e)
}
