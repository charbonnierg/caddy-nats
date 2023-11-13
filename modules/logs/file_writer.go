package logs

import (
	"io"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/logging"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	caddy.RegisterModule(FileWriter{})
}

type FileWriter struct {
	logging.FileWriter
}

func (FileWriter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.logging.writers.unique_file",
		New: func() caddy.Module { return new(FileWriter) },
	}
}

func (fw *FileWriter) Provision(ctx caddy.Context) error {
	if err := fw.FileWriter.Provision(ctx); err != nil {
		return err
	}
	return nil
}

func (fw FileWriter) String() string {
	return fw.FileWriter.String()
}

func (fw FileWriter) WriterKey() string {
	return fw.FileWriter.WriterKey()
}

func (fw FileWriter) OpenWriter() (io.WriteCloser, error) {
	// roll log files by default
	if fw.Roll == nil || *fw.Roll {
		if fw.RollSizeMB == 0 {
			fw.RollSizeMB = 100
		}
		if fw.RollCompress == nil {
			compress := true
			fw.RollCompress = &compress
		}
		if fw.RollKeep == 0 {
			fw.RollKeep = 10
		}
		if fw.RollKeepDays == 0 {
			fw.RollKeepDays = 90
		}

		l := &lumberjack.Logger{
			Filename:   fw.Filename,
			MaxSize:    fw.RollSizeMB,
			MaxAge:     fw.RollKeepDays,
			MaxBackups: fw.RollKeep,
			LocalTime:  fw.RollLocalTime,
			Compress:   *fw.RollCompress,
		}
		if err := l.Rotate(); err != nil {
			return nil, err
		}
		return l, nil
	}

	// otherwise just open a regular file
	return os.OpenFile(fw.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o666)
}

func (fw *FileWriter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return fw.FileWriter.UnmarshalCaddyfile(d)
}
