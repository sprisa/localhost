package util

import (
	"fmt"

	"github.com/rs/zerolog"
)

var Log = zerolog.New(zerolog.NewConsoleWriter())


func WrapError(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("%s: %w", msg, err)
}
