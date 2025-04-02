package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
)

type ColoredWriter struct {
	writer io.Writer
	color  string
}

func NewColoredWriter(w io.Writer, color string) *ColoredWriter {
	return &ColoredWriter{
		color:  color,
		writer: w,
	}
}

func (cw *ColoredWriter) Write(p []byte) (int, error) {
	_, err := fmt.Fprintf(cw.writer, "%s%s%s", cw.color, p, ColorReset)
	return len(p), err
}

var MediaLogger = log.New(NewColoredWriter(os.Stdout, ColorGreen), "[Media] ", log.Ldate|log.Ltime|log.Lshortfile)
var EventsLogger = log.New(NewColoredWriter(os.Stdout, ColorBlue), "[Events] ", log.Ldate|log.Ltime|log.Lshortfile)
