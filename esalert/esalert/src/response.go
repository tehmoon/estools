package main

import (
	"os/exec"
	"path/filepath"
	"io"
	"github.com/tehmoon/errors"
	"time"
	"encoding/json"
)

type Response struct {
	Action string `json:"action"`
	Args []string `json:"args"`
	ExpireAt time.Time`json:"expire_at"`
	GeneratedAt time.Time `json:"generated_at"`
	Id string `json:"id"`
	handedOffStart bool
	handedOffStop bool
	execStart bool
	execStop bool
}

func parseResponses(reader io.ReadCloser) ([]*Response, error) {
	var (
		responses []*Response
		decoder = json.NewDecoder(reader)
	)

	err := decoder.Decode(&responses)
	if err != nil {
		reader.Close()
		return nil, errors.Wrap(err, "Error decoding JSON to []*responses")
	}

	return responses, nil
}

func (r Response) Expired() (bool) {
	if time.Now().UnixNano() > r.ExpireAt.UnixNano() {
		return true
	}

	return false
}

func (r *Response) Started() {
	r.execStart = true
}

func (r *Response) Stopped() {
	r.execStop = true
}

func (r Response) Execute(base string, t NextResponse) (error) {
	bin := filepath.Join(base, r.Action)

	err := validateBin(bin)
	if err != nil {
		return errors.Wrapf(err, "Error validating ACLs for path %q", bin)
	}

	args := make([]string, 0)

	switch t {
		case NextResponseStart:
			args = append(args, "start")
		case NextResponseStop:
			args = append(args, "stop")
	}

	args = append(args, r.Args...)

	cmd := exec.Command(bin, args...)
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "Error executing command")
	}

	return nil
}
