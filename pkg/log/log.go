package pkg_log

import (
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

const Default = "default"

type LogConfig struct {
	IsProduction bool
	LogFileName  string
	Fields       map[string]interface{}
}

type LogOption func(*LogConfig)

func IsProduction(isProd bool) LogOption {
	return func(o *LogConfig) {
		o.IsProduction = isProd
	}
}

func LogName(logname string) LogOption {
	return func(o *LogConfig) {
		o.LogFileName = logname
	}
}

func LogAdditionalFields(fields map[string]interface{}) LogOption {
	return func(o *LogConfig) {
		o.Fields = fields
	}
}

func New(logOptions ...LogOption) *logrus.Logger {
	var level logrus.Level
	logger := logrus.New()

	//default configuration
	lc := &LogConfig{}
	lc.LogFileName = Default

	for _, opt := range logOptions {
		opt(lc)
	}

	// if it is production will output warn and error level
	if lc.IsProduction {
		level = logrus.WarnLevel
	} else {
		level = logrus.TraceLevel
	}

	logger.SetLevel(level)
	logger.SetOutput(colorable.NewColorableStdout())
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		//PrettyPrint:     true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcname := s[len(s)-1]
			_, filename := path.Split(f.File)
			return funcname, filename
		},
	})

	if lc.IsProduction {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcname := s[len(s)-1]
				_, filename := path.Split(f.File)
				return funcname, filename
			},
		})
	}

	if !lc.IsProduction {

		dt := time.Now().UTC()
		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   "app_logs/" + dt.Format("20060102") + "_" + lc.LogFileName,
			MaxSize:    50, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
			Level:      logrus.TraceLevel,
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
				CallerPrettyfier: func(f *runtime.Frame) (string, string) {
					s := strings.Split(f.Function, ".")
					funcname := s[len(s)-1]
					_, filename := path.Split(f.File)
					return funcname, filename
				},
			},
		})

		if err != nil {
			logger.Fatalf("Failed to initialize file rotate hook: %v", err)
		}

		logger.AddHook(rotateFileHook)
	}
	logger.AddHook(&DefaultFieldHook{lc.Fields})
	return logger

}
