package modes

import (
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/julienschmidt/httprouter"
	"github.com/rpoletaev/supervisord/process"
)

// AddVDEMasterHandlers add specific handlers to router
func (mode *VDEMaster) registerHandlers(router *httprouter.Router) {
	router.Handle(http.MethodPost, "/vde/create", mode.createVDEHandler)
	router.Handle(http.MethodPost, "/vde/start", mode.startVDEHandler)
	router.Handle(http.MethodPost, "/vde/stop", mode.stopVDEHandler)
	router.Handle(http.MethodPost, "/vde/delete", mode.deleteVDEHandler)
	router.Handle(http.MethodGet, "/vde", mode.listVDEHandler)
}

func (mode *VDEMaster) createVDEHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := r.FormValue("name")
	if len(name) == 0 {
		http.Error(w, "name is empty", http.StatusBadRequest)
		return
	}

	user := r.FormValue("dbUser")
	password := r.FormValue("dbPassword")

	if err := mode.CreateVDE(name, user, password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "VDE '%s' created", name)
}

func (mode *VDEMaster) startVDEHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := r.FormValue("name")
	if len(name) == 0 {
		http.Error(w, "name is empty", http.StatusBadRequest)
		return
	}

	proc := mode.processes.Find(name)
	if proc == nil {
		http.Error(w, fmt.Sprintf("process '%s' not found", name), http.StatusNotFound)
		return
	}

	state := proc.GetState()
	if state == process.STOPPED ||
		state == process.EXITED ||
		state == process.FATAL {
		proc.Start(true)
		fmt.Fprintf(w, "VDE '%s' is started", name)
		return
	}

	http.Error(w, fmt.Sprintf("VDE '%s' is %s", name, state), http.StatusBadRequest)
}

func (mode *VDEMaster) stopVDEHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := r.FormValue("name")
	if len(name) == 0 {
		http.Error(w, "name is empty", http.StatusBadRequest)
		return
	}

	proc := mode.processes.Find(name)
	if proc == nil {
		http.Error(w, fmt.Sprintf("process '%s' not found", name), http.StatusNotFound)
		return
	}

	state := proc.GetState()
	if state == process.RUNNING ||
		state == process.STARTING {
		proc.Stop(true)
		fmt.Fprintf(w, "VDE '%s' is stoped", name)
		return
	}

	http.Error(w, fmt.Sprintf("VDE '%s' is %s", name, state), http.StatusBadRequest)
}

func (mode *VDEMaster) deleteVDEHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := r.FormValue("name")
	if len(name) == 0 {
		http.Error(w, "name is empty", http.StatusBadRequest)
		return
	}

	proc := mode.processes.Find(name)
	if proc == nil {
		http.Error(w, fmt.Sprintf("process '%s' not found", name), http.StatusNotFound)
		return
	}

	proc.Stop(true)

}

func (mode *VDEMaster) listVDEHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	enc := json.NewEncoder(w)
	enc.Encode(mode.ListProcess())
}
