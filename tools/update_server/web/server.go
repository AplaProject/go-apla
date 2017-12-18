package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	version "github.com/hashicorp/go-version"

	"github.com/AplaProject/go-apla/tools/update_client/structs"
	"github.com/AplaProject/go-apla/tools/update_server/config"
	"github.com/AplaProject/go-apla/tools/update_server/storage"
	"github.com/AplaProject/go-apla/tools/update_server/web/middleware"
)

// Server is storing web dependencies
type Server struct {
	Db   storage.Engine
	Conf *config.Config
}

// Run is running web server
func (s *Server) Run() error {
	return http.ListenAndServe(s.Conf.Host+":"+s.Conf.Port, s.GetRoutes())
}

// GetRoutes returning all web server routes
func (s *Server) GetRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/last", s.getLastVersion)
		r.Get("/version", s.getVersions)

		r.Route("/binary", func(r chi.Router) {
			r.Use(middleware.Auth(s.Conf.Login, s.Conf.Pass))

			r.Post("/", s.addBinary)

			r.Route("/{version}", func(r chi.Router) {
				r.Delete("/", s.removeBinary)
				r.Get("/{GOOS}/{GOARCH}", s.getBinary)
			})
		})
	})

	return r
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

func (s *Server) getLastVersion(w http.ResponseWriter, r *http.Request) {
	versions, err := s.Db.GetVersionsList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	version, err := getLast(versions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	binary, err := s.Db.GetBinary(version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(binary)
}

func (s *Server) getVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := s.Db.GetVersionsList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, version := range versions {
		w.Write([]byte(version + "|"))
	}
}

func (s *Server) getBinary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars["version"] + "_" + vars["GOOS"] + "_" + vars["GOARCH"]
	binary, err := s.Db.GetBinary(version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(binary)
}

func (s *Server) addBinary(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request structs.Request
	err = json.Unmarshal(body, &request)
	if request.CheckLogin(s.Conf.Login, s.Conf.Pass) != true {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	public, err := os.Open(s.Conf.PubkeyPath)
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
	err = s.Db.AddBinary(data, request.B.Version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) removeBinary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars["version"]

	err := s.Db.DeleteBinary(version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
