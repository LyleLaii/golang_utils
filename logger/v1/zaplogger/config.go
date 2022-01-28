package zaplogger

import "github.com/pkg/errors"

const RunModeFlagName = "running.mode"
const RunModeFlagHelp = "Running Mode,default is release. One of: [dev, debug, release]"

const LevelFlagName = "log.level"
const LevelFlagHelp = "Log level, default is info. One of: [debug, info, warn, error]"

const LogMaxBackupsFlagName = "log.max-backups"
const LogMaxBackupsFlagHelp = "The maximum number of old log files to retain. Default is 5."

const LogMaxDaysFlagName = "log.max-days"
const LogMaxDaysFlagHelp = "The maximum number of days to retain old log files. Default is 30"

type Level int

const (
	Debug Level = iota + 1
	Info
	Warn
	Error
	Panic
)

type LogLevel struct {
	s string
	l Level
}

func (l *LogLevel) Level() Level {
	return l.l
}

func (l *LogLevel) String() string {
	return l.s
}

func (l *LogLevel) Set(s string) error {
	l.s = s
	switch s {
	case "debug":
		l.l = Debug
	case "info":
		l.l = Info
	case "warn":
		l.l = Warn
	case "error":
		l.l = Error
	default:
		return errors.Errorf("unrecognized log level %q", s)
	}
	return nil
}

type RunMode struct {
	s string
}

func (r *RunMode) String() string {
	return r.s
}

func (r *RunMode) Set(s string) error {
	switch s {
	case "dev", "debug", "release":
		r.s = s
	default:
		return errors.Errorf("unrecognized running mode %q", s)
	}
	return nil
}

type RunningConfig struct {
	Level      *LogLevel
	RunMode    *RunMode
	MaxBackups int
	MaxDays    int
}

func NewRunConf(level, mode string, maxdays, maxbackups int) *RunningConfig {
	c := &RunningConfig{}
	c.MaxDays = maxdays
	c.MaxBackups = maxbackups
	c.Level = &LogLevel{}
	c.Level.Set(level)
	c.RunMode = &RunMode{}
	c.RunMode.Set(mode)
	return c
}
