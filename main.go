package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type listArtist struct {
	Id           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
}
type listLocation struct {
	Id        int
	Locations []string `json:"locations"`
}

type listDates struct {
	Dates []string `json:"dates"`
}
type listRelations struct {
	DatesLocations map[string][]string `json:"datesLocations"`
}
type RelationsStruct struct {
	Location string
	Date     string
}
type listAll struct {
	Artists   listArtist
	Locations listLocation
	Dates     listDates
	Relations listRelations
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/groupe/", groupeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	server := &http.Server{
		Addr:              ":8080",           //adresse du server (le port choisi est à titre d'exemple)
		ReadHeaderTimeout: 10 * time.Second,  // temps autorisé pour lire les headers
		WriteTimeout:      10 * time.Second,  // temps maximum d'écriture de la réponse
		IdleTimeout:       120 * time.Second, // temps maximum entre deux rêquetes
		MaxHeaderBytes:    1 << 20,           // 1 MB // maxinmum de bytes que le serveur va lire
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
func fetchData(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("bad request:%v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected status code:%d", resp.StatusCode)
		return fmt.Errorf("expected status code:%d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		log.Printf("json decode error:%v", err)
	}
	return nil
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data map[string]interface{}) {
	page, err := template.ParseFiles("template/" + tmpl + ".html")
	if err != nil {
		w.WriteHeader(404)
		http.Error(w, "error 404", http.StatusNotFound)
		log.Printf("error template %v", err)
		return
	}
	err = page.Execute(w, data)
	if err != nil {
		http.Error(w, "Error 500, Internal server error", http.StatusInternalServerError)
		log.Printf("error template %v", err)
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderTemplate(w, "error404", nil)
	} else {
		var apiResponse []listArtist
		err := fetchData("https://groupietrackers.herokuapp.com/api/artists", &apiResponse)
		if err != nil {
			http.Error(w, "Error fetching data", http.StatusInternalServerError)
			log.Printf("error fetching data: %v", err)
			return
		}
		data := map[string]interface{}{
			"apiResponse": apiResponse,
		}
		RenderTemplate(w, "index", data)

	}
}

func groupeHandler(w http.ResponseWriter, r *http.Request) {
	Path := strings.Split(r.URL.Path, "/")
	if len(Path) < 3 {
		RenderTemplate(w, "error404", nil)
		return
	}
	GroupeId, err := strconv.Atoi(Path[2])
	if GroupeId < 0 || GroupeId > 52 {
		RenderTemplate(w, "error404", nil)
		return
	}
	if err != nil {
		RenderTemplate(w, "error404", nil)
		return
	}
	var apiResponse listAll
	var artist listArtist
	var dates listDates
	var locations listLocation

	GroupeURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/artists/%d", GroupeId)

	err = fetchData(GroupeURL, &apiResponse.Artists)
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		log.Printf("error fetching data: %v", err)
		return
	}

	LocationURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%d", GroupeId)

	err = fetchData(LocationURL, &apiResponse.Locations)
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		log.Printf("error fetching data: %v", err)
		return
	}

	ConcertDatesURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/dates/%d", GroupeId)

	err = fetchData(ConcertDatesURL, &apiResponse.Dates)
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		log.Printf("error fetching data: %v", err)
		return
	}
	RelationsURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", GroupeId)

	err = fetchData(RelationsURL, &apiResponse.Relations)
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		log.Printf("error fetching data: %v", err)
		return
	}
	err_artist := fetchData(GroupeURL, &artist)
	err_locations := fetchData(LocationURL, &locations)
	err_dates := fetchData(ConcertDatesURL, &dates)
	if err_artist != nil {
		log.Println(err_artist)
		RenderTemplate(w, "error400", nil)
	}
	if err_locations != nil {
		log.Println(err_locations)
		RenderTemplate(w, "error400", nil)
	}
	if err_dates != nil {
		log.Println(err_dates)
		RenderTemplate(w, "error400", nil)
	}
	var relations []RelationsStruct
	for i := 0; i < len(locations.Locations); i++ {
		relation := RelationsStruct{
			Location: Capitalize(strings.ReplaceAll(strings.ReplaceAll(locations.Locations[i], "-", " - "), "_", " ")),
			Date:     strings.ReplaceAll(strings.ReplaceAll(dates.Dates[i], "*", ""), "-", "/"),
		}
		relations = append(relations, relation)
	}
	data := map[string]interface{}{
		"Artists":   artist,
		"Relations": relations,
	}
	RenderTemplate(w, "groupe", data)
}
func Capitalize(s string) string {
	var result string
	IsNewWord := true
	for _, l := range s {
		alph := (l >= 'a' && l <= 'z') || (l >= 'A' && l <= 'Z') || (l >= '0' && l <= '9')
		if alph {
			if IsNewWord {
				if l >= 'a' && l <= 'z' {
					l = l + -32
				}
				IsNewWord = false
			} else {
				if l >= 'A' && l <= 'Z' {
					l = l + 32
				}
			}
		} else {
			IsNewWord = true
		}
		result += string(l)
	}
	return result
}
func Error(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	var tmpl string
	switch status {
	case http.StatusBadRequest:
		tmpl = "error400"
	case http.StatusNotFound:
		tmpl = "error404"
	case http.StatusInternalServerError:
		tmpl = "error500"
	}
	page, err := template.ParseFiles("templates/" + tmpl + ".html")
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}
	data := struct {
		Status  int
		Message string
	}{
		Status:  status,
		Message: message,
	}
	err = page.Execute(w, data)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
	}
}
