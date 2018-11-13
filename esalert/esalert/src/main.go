package main

import (
	"net/url"
	"github.com/tehmoon/errors"
	"time"
	"log"
	"net/http"
	"path"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
}

func main() {
	flags := parseFlags()

	cr := newCacheResponses()

	go func() {
		for ;; time.Sleep(time.Second) {
			for response := cr.Next(NextResponseStop); response != nil; response = cr.Next(NextResponseStop) {
				err := response.Execute(flags.ActionsDir, NextResponseStop)
				if err != nil {
					continue
				}

				response.Stopped()
			}
		}
	}()

	go func() {
		for ;; time.Sleep(time.Second) {
			for response := cr.Next(NextResponseStart); response != nil; response = cr.Next(NextResponseStart) {
				err := response.Execute(flags.ActionsDir, NextResponseStart)
				if err != nil {
					continue
				}

				response.Started()
			}
		}
	}()

	for ;; time.Sleep(5 * time.Second) {
		for _, tag := range flags.Tags {
			responses, err := fetchTag(flags.Url, tag)
			if err != nil {
				log.Println(errors.Wrapf(err, "Error fetch tag %q", tag))
				continue
			}

			cr.Update(responses)
		}
	}
}

func fetchTag(base *url.URL, tag string) ([]*Response, error) {
	uri, err := url.Parse(path.Join(base.Path, "response", "tag", tag))
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing uri")
	}

	endpoint := base.ResolveReference(uri).String()

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "Error calling endpoint %q", endpoint)
	}

	if resp.StatusCode != 200 {
		return nil, errors.Wrapf(err, "Bad status code %d at %q", resp.StatusCode, endpoint)
	}

	responses, err := parseResponses(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing body to responses on endpoint %q", endpoint)
	}
	defer resp.Body.Close()

	return responses, nil
}
