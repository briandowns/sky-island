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

// functionRunHandler handles requests to run functions
func (h *handler) functionRunHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		}

		clonePath := h.conf.Jails.BaseJailDir + buildJailSrcDirPath
		if !utils.Exists(clonePath + req.URL) {
			h.logger.Log("msg", "cloning "+req.URL)
			h.rsvc.CloneRepo(clonePath, req.URL)
		}

		importElems := strings.Split(req.URL, "/")
		td := &tmplData{
			PKGName:    importElems[len(importElems)-1],
			ImportPath: req.URL,
			Call:       req.Call,
		}
		t, err := template.New(req.URL).Parse(mainTmpl)
		if err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		cmdDir := fmt.Sprintf(cmdDirPath, h.conf.Jails.BaseJailDir, req.URL)
		if !utils.Exists(cmdDir) {
			if err := os.Mkdir(cmdDir, os.ModePerm); err != nil {
				h.logger.Log("error", err.Error())
			}
		}

		mainFile := fmt.Sprintf(mainFilePath, h.conf.Jails.BaseJailDir, req.URL)
		code, err := os.Create(mainFile)
		if err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		defer code.Close()

		if err = t.Execute(code, td); err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}

		buildCommand := []string{jailGoBin, "build", "-o", "/tmp/" + id, "-v", req.URL + "/cmd"}

		fullBuildArgs := []string{"-c", "-n", id, "ip4=disable", "path=" + h.conf.Jails.BaseJailDir + "/build", "host.hostname=build", "mount.devfs"}
		fullBuildArgs = append(fullBuildArgs, buildCommand...)
		buildCmd := exec.Command("jail", fullBuildArgs...)
		buildResult, err := buildCmd.CombinedOutput()
		if err != nil {
			h.logger.Log("error", string(buildResult))
			h.ren.JSON(w, 500, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}

		src := filepath.Join(h.conf.Jails.BaseJailDir, "build/tmp", id)
		dst := filepath.Join(h.conf.Jails.BaseJailDir, id, "tmp", id)
		if err := copyBinary(dst, src); err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}

		cm := strconv.Itoa(h.conf.Jails.ChildrenMax)
		funcExecArgs := []string{"-c", "-n", id, "children.max=" + cm, "path=" + h.conf.Jails.BaseJailDir + "/" + id, "host.hostname=" + id, "mount.devfs"}
		if req.IP4 {
			ip, err := h.networksvc.Allocate()
			if err != nil {
				h.logger.Log("error", err.Error())
				h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
				return
			}
			h.logger.Log("msg", "received ip allocation: "+ip)
			funcExecArgs = append(funcExecArgs, "interface="+h.conf.Network.IP4.Interface, "ip4=new", "ip4.addr="+ip)
		} else {
			funcExecArgs = append(funcExecArgs, "ip4=disable")
		}
		funcExecArgs = append(funcExecArgs, "command=/tmp/"+id)
		execRes, err := exec.Command("jail", funcExecArgs...).Output()
		if err != nil {
			h.logger.Log("error", err.Error())
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		h.ren.JSON(w, http.StatusOK, functionRunResponse{Timestamp: time.Now().UTC().Unix(), Data: string(execRes)})
		h.metrics.Histogram("handlers.function.run", 1)
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
