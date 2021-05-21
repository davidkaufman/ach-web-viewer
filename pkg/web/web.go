package web

import (
	"fmt"
	"net/http"
	"strings"

	webdisplay "github.com/moov-io/ach-web-viewer/pkg/display/web"
	"github.com/moov-io/ach-web-viewer/pkg/filelist"
	"github.com/moov-io/ach-web-viewer/pkg/service"
	"github.com/moov-io/base/log"

	"github.com/gorilla/mux"
)

func AppendRoutes(env *service.Environment, listers filelist.Listers, basePath string) {
	env.PublicRouter.Methods("GET").Path("/").HandlerFunc(listFiles(env.Logger, listers, basePath))
	env.PublicRouter.Methods("GET").PathPrefix("/sources/{sourceID}/").HandlerFunc(getFile(env.Logger, env.Config.Display, listers, basePath))
}

func listFiles(logger log.Logger, listers filelist.Listers, basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := listers.GetFiles()
		if err != nil {
			logger.Set("service", log.String("web")).Error().LogErrorf("problem listing files: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/html")

		for _, files := range resp {
			fmt.Fprintf(w, "<strong>%s</strong> (%s)", files.SourceID, files.SourceType)
			for i := range files.Files {
				fullName := fmt.Sprintf("%s%s", files.Files[i].StoragePath, files.Files[i].Name)
				fmt.Fprintf(w, `<br /><a href="%s/sources/%s/%s">%s</a>`, basePath, files.SourceID, fullName, files.Files[i].Name)
			}
			fmt.Fprint(w, "<br /><br />")
		}
	}
}

func getFile(logger log.Logger, cfg service.DisplayConfig, listers filelist.Listers, basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sourceID := mux.Vars(r)["sourceID"]
		path := strings.TrimPrefix(r.URL.Path, fmt.Sprintf("%s/sources/%s/", basePath, sourceID))

		file, err := listers.GetFile(sourceID, path)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		webdisplay.File(w, &cfg, file)
	}
}
