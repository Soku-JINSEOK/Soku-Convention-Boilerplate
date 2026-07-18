package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type envelope struct {
	OK      bool          `json:"ok"`
	Command string        `json:"command"`
	Error   *errorPayload `json:"error"`
	Data    any           `json:"data"`
}

type output struct {
	stdout io.Writer
	stderr io.Writer
	json   bool
	wrote  bool
}

func (o *output) help(command, help string) error {
	if o.json {
		return o.writeEnvelope(envelope{
			OK:      true,
			Command: command,
			Data: struct {
				Help string `json:"help"`
			}{Help: help},
		})
	}
	_, err := fmt.Fprint(o.stdout, help)
	o.wrote = true
	return err
}

func (o *output) version(metadata BuildMetadata) error {
	if o.json {
		return o.writeEnvelope(envelope{
			OK:      true,
			Command: "version",
			Data: struct {
				Version string `json:"version"`
				Commit  string `json:"commit"`
				BuiltAt string `json:"built_at"`
			}{
				Version: metadata.Version,
				Commit:  metadata.Commit,
				BuiltAt: metadata.BuiltAt,
			},
		})
	}
	_, err := fmt.Fprintf(o.stdout, "soku %s\n", metadata.Version)
	o.wrote = true
	return err
}

func (o *output) success(command string) error {
	if !o.json {
		return nil
	}
	return o.writeEnvelope(envelope{
		OK:      true,
		Command: command,
		Data:    struct{}{},
	})
}

func (o *output) failure(command string, exitError *ExitError) {
	if o.wrote {
		return
	}
	if o.json {
		_ = o.writeEnvelope(envelope{
			OK:      false,
			Command: command,
			Error: &errorPayload{
				Code:    exitError.Key,
				Message: exitError.Error(),
			},
			Data: nil,
		})
		return
	}
	_, _ = fmt.Fprintf(o.stderr, "Error: %s\n", exitError.Error())
	o.wrote = true
}

func (o *output) writeEnvelope(value envelope) error {
	encoder := json.NewEncoder(o.stdout)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(value)
	o.wrote = true
	return err
}
