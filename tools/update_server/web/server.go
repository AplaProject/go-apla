package web

import (
	"net/http"

	"github.com/go-chi/chi"

	"fmt"

	"encoding/json"

	"bytes"

	"github.com/AplaProject/go-apla/tools/update_server/config"
	"github.com/AplaProject/go-apla/tools/update_server/crypto"
	"github.com/AplaProject/go-apla/tools/update_server/model"
	"github.com/AplaProject/go-apla/tools/update_server/storage"
	"github.com/AplaProject/go-apla/tools/update_server/web/middleware"
	"github.com/go-chi/render"
)

// Server is storing web dependencies
type Server struct {
	Db        storage.Engine
	Conf      *config.Config
	Signer    crypto.Signer
	PublicKey []byte
}

// Run is running web server
func (s *Server) Run() error {
	return http.ListenAndServe(s.Conf.Host+":"+s.Conf.Port, s.GetRoutes())
}

// GetRoutes returning all web server routes
func (s *Server) GetRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/private", func(r chi.Router) {
			r.Use(middleware.Auth(s.Conf.Login, s.Conf.Pass))

			r.Route("/binary", func(r chi.Router) {
				r.Post("/", s.addBinary)
				r.Delete("/{os}/{arch}/{version}", s.removeBinary)
			})
		})

		r.Route("/{os}/{arch}", func(r chi.Router) {
			r.Get("/last", s.getLastBuildInfo)
			r.Get("/versions", s.getVersions)

			r.Route("/{version}", func(r chi.Router) {
				r.Get("/", s.getBuildInfo)
				r.Get("/binary", s.getBuild)
			})
		})
	})

	return r
}

func (s *Server) getLastBuildInfo(w http.ResponseWriter, r *http.Request) {
	versions, err := s.Db.GetVersionsList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	os := chi.URLParam(r, "os")
	a := chi.URLParam(r, "arch")

	lv, err := model.GetLastVersion(versions, os, a)
	if err != nil {
		s.HTTPError(w, r, http.StatusBadRequest, "Wrong os/arch params")
		return
	}

	var ev model.Version
	if lv == ev {
		s.HTTPError(w, r, http.StatusNotFound, "Nothing here yet")
		return
	}

	binary, err := s.Db.Get(model.Build{Version: lv})
	if err != nil {
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}

	binary.Body = []byte{}
	s.JSON(w, r, BuildInfoResponse{Build: binary})
}

func (s *Server) getVersions(w http.ResponseWriter, r *http.Request) {
	os := chi.URLParam(r, "os")
	a := chi.URLParam(r, "arch")

	versions, err := s.Db.GetVersionsList()
	if err != nil {
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}

	s.JSON(w, r, model.VersionFilter(versions, os, a))
}

func (s *Server) getBuild(w http.ResponseWriter, r *http.Request) {
	v := chi.URLParam(r, "version")
	os := chi.URLParam(r, "os")
	a := chi.URLParam(r, "arch")
	rb := model.Build{Version: model.Version{Number: v, OS: os, Arch: a}}

	b, err := s.Db.Get(rb)
	if err != nil {
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}

	if bytes.Equal(b.Body, []byte{}) {
		s.HTTPError(w, r, http.StatusNotFound, "No such version")
		return
	}

	var ev model.Version
	if b.Version == ev {
		s.HTTPError(w, r, http.StatusNotFound, "Nothing here yet")
		return
	}

	w.Write(b.Body)
}

// BuildInfoResponse is same to model.Build but without encoding body to json (for body see getBuild action)
type BuildInfoResponse struct {
	model.Build
	Body []byte `json:"body,omitempty"`
}

func (s *Server) getBuildInfo(w http.ResponseWriter, r *http.Request) {
	v := chi.URLParam(r, "version")
	os := chi.URLParam(r, "os")
	a := chi.URLParam(r, "arch")
	rb := model.Build{Version: model.Version{Number: v, OS: os, Arch: a}}

	b, err := s.Db.Get(rb)
	if err != nil {
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}

	var ev model.Version
	if b.Version == ev {
		s.HTTPError(w, r, http.StatusNotFound, "Nothing here yet")
		return
	}

	b.Body = []byte{}
	s.JSON(w, r, BuildInfoResponse{Build: b})
}

func (s *Server) addBinary(w http.ResponseWriter, r *http.Request) {
	var b model.Build
	err := render.DecodeJSON(r.Body, &b)

	if err != nil {
		s.HTTPError(w, r, http.StatusBadRequest, "Problem with decoding json")
		return
	}

	verified, err := s.Signer.CheckSign(b, s.PublicKey)
	if err != nil || !verified {
		s.HTTPError(w, r, http.StatusBadRequest, "Wrong binary sign")
		return
	}

	if !b.Validate() {
		s.HTTPError(w, r, http.StatusBadRequest, fmt.Sprintf("Wrong os+arch, available systems list: %s", model.GetAvailableVersions()))
		return
	}

	aeb, err := s.Db.Get(b)
	if err != nil {
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}

	if aeb.Version == b.Version {
		s.HTTPError(w, r, http.StatusBadRequest, "Version already exists")
		return
	}

	err = s.Db.Add(b)
	if err != nil {
		fmt.Println(err)
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}

	s.JSON(w, r, struct{}{})
}

func (s *Server) removeBinary(w http.ResponseWriter, r *http.Request) {
	v := chi.URLParam(r, "version")
	os := chi.URLParam(r, "os")
	a := chi.URLParam(r, "arch")
	rb := model.Build{Version: model.Version{Number: v, OS: os, Arch: a}}

	err := s.Db.Delete(rb)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.JSON(w, r, struct{}{})
}

func (s *Server) HTTPError(w http.ResponseWriter, r *http.Request, status int, error string) {
	w.WriteHeader(status)
	s.JSON(w, r, error)
}

func (s *Server) JSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
