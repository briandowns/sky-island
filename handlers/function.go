package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/briandowns/sky-island/utils"
	"github.com/pborman/uuid"
)

const (
	jailGoPath        = "/root/go"
	jailGoInstallpath = "/usr/local/go/bin/go"
	jailGoBin         = "command=" + jailGoInstallpath
)

const (
	buildJailSrcDirPath = "/build/root/go/src/"
	cmdDirPath          = "%s/build/root/go/src/%s/cmd"
	mainFilePath        = "%s/build/root/go/src/%s/cmd/main.go"
)

// functionRunRequest
type functionRunRequest struct {
	URL       string `json:"url"`
	Call      string `json:"call"`
	IP4       bool   `json:"ip4"`
	CacheBust bool   `json:"cache_bust"`
}

// functionRunResponse is returned upon successful
// call to the function run endpoint
type functionRunResponse struct {
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"`
}

// build
func (h *handler) build(id, url, call string) ([]byte, error) {
	importElems := strings.Split(url, "/")
	td := &tmplData{
		PKGName:    importElems[len(importElems)-1],
		ImportPath: url,
		Call:       call,
	}
	t, err := template.New(url).Parse(mainTmpl)
	if err != nil {
		return nil, err
	}
	cmdDir := fmt.Sprintf(cmdDirPath, h.conf.Jails.BaseJailDir, url)
	if !utils.Exists(cmdDir) {
		if err := os.Mkdir(cmdDir, os.ModePerm); err != nil {
			h.logger.Log("error", err.Error())
		}
	}
	mainFile := fmt.Sprintf(mainFilePath, h.conf.Jails.BaseJailDir, url)
	code, err := os.Create(mainFile)
	if err != nil {
		return nil, err
	}
	defer code.Close()

	if err = t.Execute(code, td); err != nil {
		return nil, err
	}
	buildCommand := []string{jailGoBin, "build", "-o", "/tmp/" + id, "-v", url + "/cmd"}
	fullBuildArgs := []string{"-c", "-n", id, "ip4=disable", "path=" + h.conf.Jails.BaseJailDir + "/build", "host.hostname=build", "mount.devfs"}
	fullBuildArgs = append(fullBuildArgs, buildCommand...)
	buildCmd := exec.Command("jail", fullBuildArgs...)
	return buildCmd.CombinedOutput()
}

// execute
func (h *handler) execute(id, binPath string, ip4 bool) ([]byte, error) {
	dst := filepath.Join(h.conf.Jails.BaseJailDir, id, "tmp", id)
	if err := copyBinary(dst, binPath); err != nil {
		return nil, err
	}
	cm := strconv.Itoa(h.conf.Jails.ChildrenMax)
	funcExecArgs := []string{"-c", "-n", id, "children.max=" + cm, "path=" + h.conf.Jails.BaseJailDir + "/" + id, "host.hostname=" + id, "mount.devfs"}
	if ip4 {
		ip, err := h.networksvc.Allocate([]byte(id))
		if err != nil {
			return nil, err
		}
		h.logger.Log("msg", "received ip allocation: "+ip)
		funcExecArgs = append(funcExecArgs, "interface="+h.conf.Network.IP4.Interface, "ip4=new", "ip4.addr="+ip)
	} else {
		funcExecArgs = append(funcExecArgs, "ip4=disable")
	}
	funcExecArgs = append(funcExecArgs, "command=/tmp/"+id)
	return exec.Command("jail", funcExecArgs...).Output()

}

// functionRunHandler handles requests to run functions
func (h *handler) functionRunHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer h.metrics.Histogram("handlers.function.run", 1)
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		var req functionRunRequest
		if err := json.Unmarshal(b, &req); err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, http.StatusBadRequest, map[string]string{"error": http.StatusText(http.StatusBadRequest)})
			return
		}
		id := uuid.NewUUID().String()
		if err := h.jsvc.CreateJail(id, true); err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		defer h.jsvc.RemoveJail(id)

		if req.CacheBust {
			if err := h.rsvc.RemoveRepo(req.URL); err != nil {
				h.logger.Log("error", err.Error())
				h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
				return
			}
			h.binCache.Set(req.URL, "")
		}
		var binPath string
		binPath = h.binCache.Get(req.URL)
		if binPath != "" {
			execRes, err := h.execute(id, binPath, req.IP4)
			if err != nil {
				h.logger.Log("error", err.Error())
				h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
				return
			}
			h.ren.JSON(w, http.StatusOK, functionRunResponse{Timestamp: time.Now().UTC().Unix(), Data: string(execRes)})
			return
		}
		clonePath := h.conf.Jails.BaseJailDir + buildJailSrcDirPath
		if !utils.Exists(clonePath + req.URL) {
			h.logger.Log("msg", "cloning "+req.URL)
			if err := h.rsvc.CloneRepo(clonePath, req.URL); err != nil {
				h.logger.Log("error", err.Error())
				h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
				return
			}
		}
		buildRes, err := h.build(id, req.URL, req.Call)
		if err != nil {
			h.logger.Log("error", err.Error()+" "+string(buildRes))
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		execRes, err := h.execute(id, h.conf.Jails.BaseJailDir+"/build/tmp/"+id, req.IP4)
		if err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		h.ren.JSON(w, http.StatusOK, functionRunResponse{Timestamp: time.Now().UTC().Unix(), Data: string(execRes)})
	}
}

// tmplData contains the data passed to the tempalte
// engine to render the code for compilation
type tmplData struct {
	PKGName    string
	ImportPath string
	Call       string
}

// mainTmpl is the template used for function execution
const mainTmpl = `// generated by sky-island
// DO NOT EDIT

package main

import (
	"fmt"

	"{{.ImportPath}}"
)

func main() {
	fmt.Print({{.PKGName}}.{{.Call}})
}
`

// copyBinary copies the given src to the given destination
func copyBinary(dst, src string) error {
	bb, err := os.Open(src)
	if err != nil {
		return err
	}
	defer bb.Close()
	be, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer be.Close()
	if _, err := io.Copy(be, bb); err != nil {
		return err
	}
	if err := os.Chmod(dst, 0777); err != nil {
		return err
	}
	return nil
}
