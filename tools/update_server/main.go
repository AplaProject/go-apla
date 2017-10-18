package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"github.com/AplaProject/go-apla/tools/update_server/structs"

	"github.com/AplaProject/go-apla/tools/update_server/config"
	"github.com/AplaProject/go-apla/tools/update_server/database"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-version"
)

var (
	db   *database.Database
	conf *config.Config
)

func main() {
	conf = &config.Config{}
	err := conf.Read()
	if err != nil {
		fmt.Println("can't read config: ", err)
		return
	}
	db = &database.Database{}
	err = db.Open(conf.DBPath)
	if err != nil {
		fmt.Println("can't open database: ", err)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/binary", addBinary).Methods("POST")
	r.HandleFunc("/v1/binary/{version}/{GOOS}/{GOARCH}", getBinary).Methods("GET")
	r.HandleFunc("/v1/binary/{version}", removeBinary).Methods("DELETE")
	r.HandleFunc("/v1/last", getLastVersion).Methods("GET")

	r.HandleFunc("/v1/version", getVersions).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(conf.Host+":"+conf.Port, r)
}

func getLast(versions []string) (string, error) {
	var vers []*version.Version
	for _, v := range versions {
		t, err := version.NewVersion(v)
		if err != nil {
			return "", err
		}
		vers = append(vers, t)
	}
	sort.Sort(version.Collection(vers))
	return vers[len(vers)-1].String(), nil
}

func getLastVersion(w http.ResponseWriter, r *http.Request) {
	versions, err := db.GetVersionsList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	version, err := getLast(versions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	binary, err := db.GetBinary(version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(binary)
}

func getVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := db.GetVersionsList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, version := range versions {
		w.Write([]byte(version + "|"))
	}
}

func getBinary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars["version"] + "_" + vars["GOOS"] + "_" + vars["GOARCH"]
	binary, err := db.GetBinary(version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(binary)
}

func addBinary(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request structs.Request
	err = json.Unmarshal(body, &request)
	if request.CheckLogin(conf.Login, conf.Pass) != true {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	public, err := os.Open(conf.PubkeyPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pubData, err := ioutil.ReadAll(public)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	verified, err := request.B.CheckSign(pubData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if verified != true {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(request.B)
	fmt.Println(len(data))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.AddBinary(data, request.B.Version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func removeBinary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars["version"]

	err := db.DeleteBinary(version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
