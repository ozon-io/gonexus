package nexus

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	// "net/http/httputil"

	"github.com/hokiegeek/gonexus/iq"
	"github.com/hokiegeek/gonexus/rm"
)

/*
func iqResultToComponent(r nexusiq.ComponentEvaluationResult) component {
	var c component
	c.ComponentIdentifier.Format = r.Component.ComponentIdentifier.Format
	c.ComponentIdentifier.Coordinates.ArtifactID = r.Component.ComponentIdentifier.Coordinates.ArtifactID
	c.ComponentIdentifier.Coordinates.GroupID = r.Component.ComponentIdentifier.Coordinates.GroupID
	c.ComponentIdentifier.Coordinates.Version = r.Component.ComponentIdentifier.Coordinates.Version
	c.ComponentIdentifier.Coordinates.Extension = r.Component.ComponentIdentifier.Coordinates.Extension
	// c.Quarantined = false
	if highestViolation := r.HighestThreatPolicy(); highestViolation != nil {
		// c.HighestThreatLevel = true
		c.ThreatLevel = highestViolation.ThreatLevel
		c.PolicyName = highestViolation.PolicyName
	}
	return c
}
*/

// Server provides an HTTP wrapper with optimized for communicating with a Nexus server
type Server struct {
	host, username, password string
}

func (s *Server) http(method, endpoint string, payload io.Reader) ([]byte, *http.Response, error) {
	url := fmt.Sprintf(endpoint, s.host)
	request, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, nil, err
	}

	request.SetBasicAuth(s.username, s.password)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	// dump, _ := httputil.DumpRequest(request, true)
	// fmt.Printf("%q\n", dump)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		return body, resp, err
	}

	return nil, resp, errors.New(resp.Status)
}

// Get performs an HTTP GET against the indicated endpoint
func (s *Server) Get(endpoint string) ([]byte, *http.Response, error) {
	return s.http("GET", endpoint, nil)
}

// Post performs an HTTP POST against the indicated endpoint
func (s *Server) Post(endpoint string, payload []byte) ([]byte, *http.Response, error) {
	return s.http("POST", endpoint, bytes.NewBuffer(payload))
}

// Put performs an HTTP PUT against the indicated endpoint
func (s *Server) Put(endpoint string, payload []byte) ([]byte, *http.Response, error) {
	return s.http("PUT", endpoint, bytes.NewBuffer(payload))
}

// Del performs an HTTP DELETE against the indicated endpoint
func (s *Server) Del(endpoint string) error {
	_, _, err := s.http("DELETE", endpoint, nil)
	return err
}

// RmItemToIQComponent converts a repo item to an IQ component
func RmItemToIQComponent(rm nexusrm.RepositoryItem) nexusiq.Component {
	var iqc nexusiq.Component
	switch rm.Format {
	case "maven2":
		iqc.ComponentIdentifier.Format = "maven"
		iqc.ComponentIdentifier.Coordinates.Extension = "jar"
	case "rubygems":
		iqc.ComponentIdentifier.Format = "gem"
	case "npm":
		iqc.ComponentIdentifier.Format = "npm"
		iqc.ComponentIdentifier.Coordinates.Extension = "tgz"
	case "pipy":
		iqc.ComponentIdentifier.Format = "pypi"
	default:
		iqc.ComponentIdentifier.Format = rm.Format
	}
	iqc.ComponentIdentifier.Coordinates.ArtifactID = rm.Name
	iqc.ComponentIdentifier.Coordinates.GroupID = rm.Group
	iqc.ComponentIdentifier.Coordinates.Version = rm.Version
	iqc.Hash = rm.Hash()
	return iqc
}
