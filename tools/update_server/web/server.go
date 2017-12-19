package web

import (
	"net/http"

	"github.com/go-chi/chi"

	"fmt"

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
	Signer    crypto.BuildSigner
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
			r.Get("/last", s.getLastVersion)
			r.Get("/versions", s.getVersions)
			r.Get("/{version}", s.getBinary)
		})
	})

	return r
}

func (s *Server) getLastVersion(w http.ResponseWriter, r *http.Request) {
	versions, err := s.Db.GetVersionsList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lastBuild := model.GetLastVersion(versions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	binary, err := s.Db.Get(lastBuild)
	if err != nil {
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}
	w.Write(binary.Body)
}

func (s *Server) getVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := s.Db.GetVersionsList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, versions)
}

func (s *Server) getBinary(w http.ResponseWriter, r *http.Request) {
	v := chi.URLParam(r, "version")
	os := chi.URLParam(r, "os")
	a := chi.URLParam(r, "arch")
	rb := model.Build{Version: v, OS: os, Arch: a}

	binary, err := s.Db.Get(rb)
	if err != nil {
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}

	w.Write(binary.Body)
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

	if !b.ValidateSystem() {
		s.HTTPError(w, r, http.StatusBadRequest, fmt.Sprintf("Wrong os+arch, available systems list: %s", b.GetAvailableSystems()))
		return
	}

	err = s.Db.Add(b)
	if err != nil {
		s.HTTPError(w, r, http.StatusInternalServerError, "Database problems")
		return
	}

	render.JSON(w, r, struct{}{})
}

func (s *Server) removeBinary(w http.ResponseWriter, r *http.Request) {
	v := chi.URLParam(r, "version")
	os := chi.URLParam(r, "os")
	a := chi.URLParam(r, "arch")
	rb := model.Build{Version: v, OS: os, Arch: a}

	err := s.Db.Delete(rb)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, struct{}{})
}

func (s *Server) HTTPError(w http.ResponseWriter, r *http.Request, status int, error string) {
	render.Status(r, status)
	render.JSON(w, r, error)
}
