package picast

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/antage/eventsource"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	//"strconv"
	//"flag"
	"time"
)

const (
	fileserverPort = "8080"
	ssePort        = "8081"
	restPort       = "8082"
)

func RunServer(displayUpdates chan PlaylistEntry) {

	MainMedia = Media{
		Metadata:     &PlaylistEntry{},
		MediaChanged: make(chan bool),
		MediaAdded:   make(chan bool),
	}

	log.Println("Server Started.")

	// REST handler

	handler := rest.ResourceHandler{
		PreRoutingMiddlewares: []rest.Middleware{
			&MyCorsMiddleware{},
		},
	}
	handler.SetRoutes(
		/*&rest.Route{"GET", "/api/entries", api.GetAllEntries},
		&rest.Route{"GET", "/api/entries/:id", api.GetEntry},
		&rest.Route{"POST", "/api/entries", api.PostEntry},
		&rest.Route{"DELETE", "/api/entries/:id", api.DeleteEntry},

		&rest.Route{"POST", "/api/playall", api.PlayAll},
		&rest.Route{"POST", "/api/prev", api.Prev},
		&rest.Route{"POST", "/api/next", api.Next},*/

		&rest.Route{"GET", "/media/play", MainMedia.Play},
		&rest.Route{"POST", "/media/add", MainMedia.Add},
		&rest.Route{"GET", "/media/pause", MainMedia.TogglePause},
		&rest.Route{"GET", "/media/stop", MainMedia.Stop},
		&rest.Route{"GET", "/media/status", MainMedia.Status},
		&rest.Route{"GET", "/media/playlist", MainMedia.GetPlaylist},
	)

	go http.ListenAndServe(":"+restPort, &handler)

	// Server Sent Events Handler

	es := eventsource.New(
		eventsource.DefaultSettings(),
		func(req *http.Request) [][]byte {
			return [][]byte{
				[]byte("Access-Control-Allow-Origin: *"),
			}
		},
	)
	//defer es.Close()
	http.Handle("/events", es)
	go func() {
		currentState := 0

		for {
			select {
			case <-MainMedia.MediaChanged:
				displayUpdates <- MainMedia.Playlist[MainMedia.CurrentIndex]
				log.Println("Sent update to display.")
			case <-MainMedia.MediaAdded:
				es.SendEventMessage("mediaAdded", "playlistChanged", "")
			default:
				if MainMedia.Player != nil && currentState != MainMedia.Player.StatusCode() {

					switch MainMedia.Player.StatusCode() {
					case 0:
						es.SendEventMessage("stopped", "playerStateChanged", "")
					case 2:
						es.SendEventMessage("paused", "playerStateChanged", "")
					case 3:
						es.SendEventMessage("playing", "playerStateChanged", "")
						log.Println("Media playing.")
					}

					currentState = MainMedia.Player.StatusCode()
				}
			}

			time.Sleep(250 * time.Millisecond)

		}
	}()
	go http.ListenAndServe(":"+ssePort, nil)

	// HTTP handler

	// Write server configuration to js file that will be referenced by webapp
	ip, err := externalIP()
	if err != nil {
		log.Println(err)
	}

	serverConfigjs := "define(function(require, exports, module) {" + "\n\tmodule.exports = ['" + ip + "'];" + "\n});"

	err = ioutil.WriteFile("static/src/serverConfig.js", []byte(serverConfigjs), 0644)
	if err != nil {
		log.Panicln(err)
	}

	r := mux.NewRouter()

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	http.Handle("/", r)
	log.Panic(http.ListenAndServe(":"+fileserverPort, nil))
}

type MyCorsMiddleware struct{}

func (mw *MyCorsMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(writer rest.ResponseWriter, request *rest.Request) {

		corsInfo := request.GetCorsInfo()

		// Be nice with non CORS requests, continue
		// Alternatively, you may also chose to only allow CORS requests, and return an error.
		if !corsInfo.IsCors {
			// continure, execute the wrapped middleware
			handler(writer, request)
			return
		}

		// Validate the Origin
		// More sophisticated validations can be implemented, regexps, DB lookups, ...
		/*myIp, _ := externalIP()
		if corsInfo.Origin != myIp+":"+fileserverPort && corsInfo.Origin != "raspberrypi.local:"+fileserverPort {
			rest.Error(writer, "Invalid Origin", http.StatusForbidden)
			return
		}*/

		if corsInfo.IsPreflight {
			// check the request methods
			allowedMethods := map[string]bool{
				"GET":  true,
				"POST": true,
				// don't allow DELETE, for instance
			}
			if !allowedMethods[corsInfo.AccessControlRequestMethod] {
				rest.Error(writer, "Invalid Preflight Request", http.StatusForbidden)
				return
			}
			// check the request headers
			allowedHeaders := map[string]bool{
				"Accept":       true,
				"Content-Type": true,
			}
			for _, requestedHeader := range corsInfo.AccessControlRequestHeaders {
				if !allowedHeaders[requestedHeader] {
					rest.Error(writer, "Invalid Preflight Request", http.StatusForbidden)
					return
				}
			}

			for allowedMethod, _ := range allowedMethods {
				writer.Header().Add("Access-Control-Allow-Methods", allowedMethod)
			}
			for allowedHeader, _ := range allowedHeaders {
				writer.Header().Add("Access-Control-Allow-Headers", allowedHeader)
			}
			writer.Header().Set("Access-Control-Allow-Origin", corsInfo.Origin)
			writer.Header().Set("Access-Control-Allow-Credentials", "true")
			writer.Header().Set("Access-Control-Max-Age", "3600")
			writer.WriteHeader(http.StatusOK)
			return
		} else {
			writer.Header().Set("Access-Control-Expose-Headers", "X-Powered-By")
			writer.Header().Set("Access-Control-Allow-Origin", corsInfo.Origin)
			writer.Header().Set("Access-Control-Allow-Credentials", "true")
			// continure, execute the wrapped middleware
			handler(writer, request)
			return
		}
	}
}
