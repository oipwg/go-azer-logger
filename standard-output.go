package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func NewStandardOutput(file *os.File) OutputWriter {
	var writer = &StandardWriter{
		ColorsEnabled: true,
		Target:        file,
	}

	defaultOutputSettings := parseVerbosityLevel(os.Getenv("LOG_LEVEL"))
	writer.Settings = parsePackageSettings(os.Getenv("LOG"), defaultOutputSettings)

	return writer
}

type StandardWriter struct {
	ColorsEnabled bool
	Target        *os.File
	Settings      map[string]*OutputSettings
}

func (standardWriter StandardWriter) Init() {}

func (sw StandardWriter) Write(log *Log) {
	if sw.IsEnabled(log.Package, log.Level) {
		fmt.Fprintln(sw.Target, sw.Format(log))
	}
}

func (sw *StandardWriter) IsEnabled(logger, level string) bool {
	settings := sw.LoggerSettings(logger)

	if level == "INFO" {
		return settings.Info
	}

	if level == "ERROR" {
		return settings.Error
	}

	if level == "TIMER" {
		return settings.Timer
	}

	return false
}

func (sw *StandardWriter) LoggerSettings(p string) *OutputSettings {
	if settings, ok := sw.Settings[p]; ok {
		return settings
	}

	// If there is a "*" (Select all) setting, return that
	if settings, ok := sw.Settings["*"]; ok {
		return settings
	}

	return muted
}
func (sw *StandardWriter) Format(log *Log) string {
    return sw.PrettyFormat(log)
}

func (sw *StandardWriter) PrettyFormat(log *Log) string {
    return fmt.Sprintf("%s %s %s%s",
        time.Now().Format("15:04:05.000"),
        sw.PrettyLabel(log),
        log.Message,
        sw.PrettyAttrs(log.Attrs))
}

func (sw *StandardWriter) PrettyLabel(log *Log) string {
    label := log.Package + sw.PrettyLabelExt(log) + ":"
    if sw.ColorsEnabled {
        return fmt.Sprintf("%s%s%s", sw.getColorFor(log.Package), label, sw.resetCode())
    }
    return label
}

func (sw *StandardWriter) PrettyLabelExt(log *Log) string {
    if log.Level == "ERROR" {
        if sw.ColorsEnabled {
            return fmt.Sprintf("(%s!%s)", red, sw.resetCode())
        }
        return "(ERROR)"
    }

    if log.Level == "TIMER" {
        elapsed := fmt.Sprintf("%v", time.Duration(log.ElapsedNano))
        if sw.ColorsEnabled {
            return fmt.Sprintf("(%s%s%s)", sw.resetCode(), elapsed, sw.getColorFor(log.Package))
        }
        return fmt.Sprintf("(%s)", elapsed)
    }

    return ""
}

func (sw *StandardWriter) PrettyAttrs(attrs *Attrs) string {
    if attrs == nil {
        return ""
    }

    result := ""
    for key, val := range *attrs {
        result = fmt.Sprintf("%s %s=%v", result, key, val)
    }

    return result
}

func (sw *StandardWriter) getColorFor(packageName string) string {
    if !sw.ColorsEnabled {
        return ""
    }
    // Existing color selection logic
    return colorFor(string);
}

func (sw *StandardWriter) resetCode() string {
    if sw.ColorsEnabled {
        return reset
    }
    return ""
}

// Accepts: foo,bar,qux@timer
//          *
//          *@error
//          *@error,database@timer
func parsePackageSettings(input string, defaultOutputSettings *OutputSettings) map[string]*OutputSettings {
	all := map[string]*OutputSettings{}
	items := strings.Split(input, ",")

	for _, item := range items {
		name, verbosity := parsePackageName(item)
		if verbosity == nil {
			verbosity = defaultOutputSettings
		}

		all[name] = verbosity
	}

	return all
}

// Accepts: users
//          database@timer
//          server@error
func parsePackageName(input string) (string, *OutputSettings) {
	parsed := strings.Split(input, "@")
	name := strings.TrimSpace(parsed[0])

	if len(parsed) > 1 {
		return name, parseVerbosityLevel(parsed[1])
	}

	return name, nil
}

func parseVerbosityLevel(val string) *OutputSettings {
	val = strings.ToUpper(strings.TrimSpace(val))

	if val == "MUTE" {
		return &OutputSettings{}
	}

	s := &OutputSettings{
		Info:  true,
		Timer: true,
		Error: true,
	}

	if val == "TIMER" {
		s.Info = false
	}

	if val == "ERROR" {
		s.Info = false
		s.Timer = false
	}

	return s
}
