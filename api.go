package picast

import (
	//"fmt"
	"log"

	"database/sql"
	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/mattn/go-sqlite3"

	"net/http"
	"strconv"
)

var DBNAME = "res/playlist.db"
var TABLENAME = "playlist"
var COLNAME1 = "url"

//var COLNAME2 = "played"

func (api *Api) InitDB() {
	db, err := sql.Open("sqlite3", DBNAME)
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

	api.DB = db
	api.DB.Exec("VACUUM")
}

func (api *Api) GetAllEntries(w rest.ResponseWriter, r *rest.Request) {
	entries := make([]PlaylistEntry, 0)

	rows, err := api.DB.Query("SELECT ROWID, " + COLNAME1 + " FROM " + TABLENAME)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			log.Fatal(err)
		}
		entries = append(entries, PlaylistEntry{Id: id, Url: url})
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	w.WriteJson(&entries)
}

func (api *Api) GetEntry(w rest.ResponseWriter, r *rest.Request) {
	idParam := r.PathParam("id")
	entries := &PlaylistEntry{}

	rows, err := api.DB.Query("SELECT ROWID, " + COLNAME1 + " FROM " + TABLENAME + " WHERE ROWID = " + idParam)
	if err != nil {
		log.Fatal(err)
	}

	if rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			log.Fatal(err)
		}
		entries = &PlaylistEntry{Id: id, Url: url}
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
	entry := PlaylistEntry{}

	err := r.DecodeJsonPayload(&entry)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = api.DB.Exec("INSERT INTO " + TABLENAME + "(" + COLNAME1 + ") VALUES('" + entry.Url + "')")
	if err != nil {
		log.Fatal(err)
		return
	}

	lastInsertId, err := api.DB.Query("SELECT last_insert_rowid()")
	if err != nil {
		log.Fatal(err)
		return
	}

	if lastInsertId.Next() {
		var id int
		if err := lastInsertId.Scan(&id); err != nil {
			log.Fatal(err)
		}
		entry.Id = id
	} else {
		rest.NotFound(w, r)
		return
	}
	if err := lastInsertId.Err(); err != nil {
		log.Fatal(err)
	}

	w.WriteJson(&entry)
}

func (api *Api) DeleteEntry(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")

	_, err := api.DB.Exec("DELETE FROM " + TABLENAME + " WHERE ROWID = " + id)
	if err != nil {
		log.Fatal(err)
	}
}

func (api *Api) LocalDelete(entry PlaylistEntry) {
	_, err := api.DB.Exec("DELETE FROM " + TABLENAME + " WHERE ROWID = " + strconv.Itoa(entry.Id))
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: Implement GetFirst()
func (api *Api) GetFirst() *PlaylistEntry {
	nextEntry := &PlaylistEntry{}

	rows, err := api.DB.Query("SELECT ROWID, " + COLNAME1 + " FROM " + TABLENAME + " ORDER BY ROWID ASC LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}

	if rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			log.Fatal(err)
		}
		nextEntry = &PlaylistEntry{Id: id, Url: url}
	} else {
		// Return empty PlaylistEntry if DB has no Entry
		return &PlaylistEntry{}
	}

	return nextEntry
}

func (api *Api) GetNext() *PlaylistEntry {
	// Return empty PlaylistEntry{} if CurrentMedia Metadata is not initialized
	if *api.CurrentMedia.Metadata == (PlaylistEntry{}) {
		return &PlaylistEntry{}
	}

	nextEntry := &PlaylistEntry{}
	idParam := api.CurrentMedia.Metadata.Id

	rows, err := api.DB.Query("SELECT ROWID, " + COLNAME1 + " FROM " + TABLENAME + " WHERE ROWID > " + strconv.Itoa(idParam) + " ORDER BY ROWID ASC LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}

	if rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			log.Fatal(err)
		}
		nextEntry = &PlaylistEntry{Id: id, Url: url}
	} else {
		// Return empty PlaylistEntry if DB has no more Entries
		return &PlaylistEntry{}
	}

	return nextEntry
}
