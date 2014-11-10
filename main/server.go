package picast

import (
	"github.com/imheresamir/picast"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {

	mainMedia := picast.Media{Metadata: &PlaylistEntry{}}
	api := picast.Api{CurrentPlayer: &mainMedia}
	api.InitDB()

	log.Println("Server Started.")

	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
	}
	handler.SetRoutes(
		&rest.Route{"GET", "/api/entries", api.GetAllEntries},
		&rest.Route{"GET", "/api/entries/:id", api.GetEntry},
		&rest.Route{"POST", "/api/entries", api.PostEntry},
		&rest.Route{"DELETE", "/api/entries/:id", api.DeleteEntry},

		&rest.Route{"POST", "/api/playall", api.PlayAll},
		&rest.Route{"POST", "/api/prev", api.Prev},
		&rest.Route{"POST", "/api/next", api.Next},

		&rest.Route{"POST", "/api/play", mainMedia.Play},
		&rest.Route{"POST", "/api/pause", mainMedia.TogglePause},
		&rest.Route{"POST", "/api/stop", mainMedia.Stop},
	)

	http.ListenAndServe(":8082", &handler)
}
