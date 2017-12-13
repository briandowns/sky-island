package jail

import (
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/briandowns/sky-island/utils"
)

// JLS holds the Go represented output from the
// jls command
type JLS struct {
	Host      string `json:"host"`
	IP4       string `json:"ip4"`
	IP6       string `json:"ip6"`
	JID       int    `json:"jid"`
	Name      string `json:"name"`
	OSRelease string `json:"OSRelease"`
	Path      string `json:"path"`
	Hostname  string `json:"hostname"`
}

// JLSRun runs the jls command to get a slice of the
// running jails
func JLSRun(w utils.Wrapper) ([]*JLS, error) {
	res, err := w.CombinedOutput("jls", "-s")
	if err != nil {
		return nil, errors.New(string(res))
	}
	var jlsData []*JLS
	for _, jd := range bytes.Split(res, []byte("\n")) {
		if len(jd) <= 1 {
			continue
		}
		jail, err := unmarshalJLS(jd)
		if err != nil {
			return nil, err
		}
		jlsData = append(jlsData, jail)
	}
	return jlsData, nil
}

// unmarshal unmarshals the data from a JLS command
// into the given struct
func unmarshalJLS(data []byte) (*JLS, error) {
	kvs := strings.Split(string(data), " ")
	var j JLS
	for _, i := range kvs {
		kv := strings.Split(i, "=")
		switch kv[0] {
		case "host":
			j.Host = kv[1]
		case "ip4":
			j.IP4 = kv[1]
		case "ip6":
			j.IP6 = kv[1]
		case "jid":
			jid, err := strconv.Atoi(kv[1])
			if err != nil {
				return nil, err
			}
			j.JID = jid
		case "name":
			j.Name = kv[1]
		case "osrelease":
			j.OSRelease = kv[1]
		case "path":
			j.Path = kv[1]
		case "host.hostname":
			j.Hostname = kv[1]
		}
	}
	return &j, nil
}
