package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"os/exec"
	"bytes"
	"strconv"
)

var DBNAME = "playlist.db"
var TABLENAME = "playlist"
var COLNAME1 = "url"
var COLNAME2 = "played"

type Api struct {
	DB *sql.DB
}

type PlaylistEntry struct {
	Id			int
	Url			string
	Played	int
}

func (api *Api) initDB() {
	db, err := sql.Open("sqlite3", DBNAME)
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

	api.DB = db
	api.DB.Exec("VACUUM")
}

func main() {
	api := Api{}
	api.initDB()

	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
	}
	handler.SetRoutes(
		rest.RouteObjectMethod("GET", "/api/entries", &api, "GetAllEntries"),
		rest.RouteObjectMethod("POST", "/api/entries", &api, "PostEntry"),
		rest.RouteObjectMethod("GET", "/api/entries/:id", &api, "GetEntry"),
		//rest.RouteObjectMethod("PUT", "/api/entries/:id", &api, "PutEntry"),
		rest.RouteObjectMethod("DELETE", "/api/entries/:id", &api, "DeleteEntry"),
	)

	http.ListenAndServe(":8082", &handler)
}

func (api *Api) GetAllEntries(w rest.ResponseWriter, r *rest.Request) {
	entries := make([]PlaylistEntry, 0)

	rows, err := api.DB.Query("SELECT ROWID, " + COLNAME1 + ", " + COLNAME2 + " FROM " + TABLENAME)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var url string
		var played int
		if err := rows.Scan(&id, &url, &played); err != nil {
			log.Fatal(err)
		}
		entries = append(entries, PlaylistEntry{Id: id, Url: url, Played: played})
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	w.WriteJson(&entries)
}

func (api *Api) GetEntry(w rest.ResponseWriter, r *rest.Request) {
	idParam := r.PathParam("id")
	entries := &PlaylistEntry{}

	rows, err := api.DB.Query("SELECT ROWID, " + COLNAME1 + ", " + COLNAME2 + " FROM " + TABLENAME + " WHERE ROWID = " + idParam)
	if err != nil {
		log.Fatal(err)
	}

	if rows.Next() {
		var id int
		var url string
		var played int
		if err := rows.Scan(&id, &url, &played); err != nil {
			log.Fatal(err)
		}
		entries = &PlaylistEntry{Id: id, Url: url, Played: played}
	} else {
		rest.NotFound(w, r)
		return
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	w.WriteJson(&entries)
}

func (api *Api) PostEntry(w rest.ResponseWriter, r *rest.Request) {
	entries := PlaylistEntry{}

	err := r.DecodeJsonPayload(&entries)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = api.DB.Exec("INSERT INTO " + TABLENAME + "(" + COLNAME1 + ") VALUES('" + entries.Url + "')")
	if err != nil {
		fmt.Println(err)
		return
	}

	lastInsertId, err := api.DB.Query("SELECT last_insert_rowid()")
	if err != nil {
		fmt.Println(err)
		return
	}

	if lastInsertId.Next() {
		var id int
		if err := lastInsertId.Scan(&id); err != nil {
			log.Fatal(err)
		}
		entries.Id = id
	} else {
		rest.NotFound(w, r)
		return
	}
	if err := lastInsertId.Err(); err != nil {
		log.Fatal(err)
	}

	go launchClient(entries.Id)
	w.WriteJson(&entries)
}

func launchClient(videoId int) {
	cmd := exec.Command("python", "client.py", "play", strconv.Itoa(videoId))
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Client output: " + out.String())
}

/*func (api *Api) PutPlaylistEntry(w rest.ResponseWriter, r *rest.Request) {
		id := r.PathParam("id")
		if self.Store[id] == nil {
				rest.NotFound(w, r)
				return
		}
		entries := PlaylistEntry{}
		err := r.DecodeJsonPayload(&entries)
		if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
		}
		entries.Id = id
		self.Store[id] = &entries
		w.WriteJson(&entries)
}*/

func (api *Api) DeleteEntry(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")

	_, err := api.DB.Exec("DELETE FROM " + TABLENAME + " WHERE ROWID = " + id)
	if err != nil {
		fmt.Println(err)
		return
	}

}
