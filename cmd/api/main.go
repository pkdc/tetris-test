package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
)

var id int

// type GRecords struct {
// 	GameRecords []GameRecord `json:"game_records"`
// }

type GameRecord struct {
	Id         int    `json:"id"`
	PlayerName string `json:"pname"`
	GameScore  string `json:"score"`
	GameTime   string `json:"time"`
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("./assets/index.html")
	if err != nil {
		http.Error(w, "Parsing Error", http.StatusInternalServerError)
		return
	}
	err = tpl.ExecuteTemplate(w, "index.html", nil)
}

func getJsonData(file *os.File, jsonRecords *[]GameRecord) {

	// get the records from record.json
	byteRecord, _ := io.ReadAll(file)
	json.Unmarshal(byteRecord, jsonRecords)
	for _, r := range *jsonRecords {
		fmt.Printf("prev records: %v\n", r)
	}
}

func recordHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Printf("-------method---------%s\n", r.Method)
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	// Get to get
	if r.Method == http.MethodGet {
		fmt.Printf("----record-GET-----\n")

		var Records []GameRecord
		f, err := os.OpenFile("record.json", os.O_RDONLY, 0644)
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "Please play the game first", http.StatusBadRequest)
		}
		// get the records from record.json
		getJsonData(f, &Records)

		js, err := json.MarshalIndent(Records, "", "\t")

		// respond with json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(js)
	}

	// Post to store
	if r.Method == http.MethodPost {
		fmt.Printf("----record-POST-----\n")
		// err := r.ParseForm()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		var recordPayload GameRecord

		err := json.NewDecoder(r.Body).Decode(&recordPayload)

		fmt.Println(recordPayload)

		// fmt.Printf("Id: %d\n", id)
		pname := recordPayload.PlayerName
		score := recordPayload.GameScore
		time := recordPayload.GameTime

		// pname := r.PostForm.Get("pname")
		fmt.Printf("Name: %s\n", pname)
		// score := r.PostForm.Get("score")
		fmt.Printf("Score: %s\n", score)
		// time := r.PostForm.Get("time")
		fmt.Printf("Time: %s\n", time)

		// try to open to read
		f, err := os.OpenFile("record.json", os.O_RDONLY, 0444)
		// if file not exist
		if errors.Is(err, fs.ErrNotExist) {
			// first record
			curRecord := GameRecord{
				Id:         1,
				PlayerName: pname,
				GameScore:  score,
				GameTime:   time,
			}

			var Records []GameRecord
			Records = append(Records, curRecord)

			for _, r := range Records {
				fmt.Printf("first record: %v\n", r)
			}

			js, err := json.MarshalIndent(Records, "", "\t")
			if err != nil {
				log.Fatal(err)
			}

			err = os.WriteFile("record.json", js, 0644)
			if err != nil {
				log.Fatal(err)
			}

			f, err := os.OpenFile("record.json", os.O_RDONLY, 0444)
			if errors.Is(err, fs.ErrNotExist) {
				log.Fatal(err)
			}
			defer f.Close()

			getJsonData(f, &Records)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(js)

			// w.Header().Set("Location", "/")
			// w.WriteHeader(http.StatusSeeOther)
			return
		}
		defer f.Close()

		// if file exist
		var Records []GameRecord
		// var recordStr string

		// get the records from record.json
		getJsonData(f, &Records)

		lastId := Records[len(Records)-1].Id

		// id is the last id + 1
		curRecord := GameRecord{
			Id:         lastId + 1,
			PlayerName: pname,
			GameScore:  score,
			GameTime:   time,
		}

		Records = append(Records, curRecord)

		fmt.Println("---------------------------------")
		for _, r := range Records {
			fmt.Printf("after records: %v\n", r)
		}

		js, err := json.MarshalIndent(Records, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile("record.json", js, 0644)
		// f.Write([]byte('\n'))

		// reply
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(js)

		// w.Header().Set("Location", "/")
		// w.WriteHeader(http.StatusSeeOther)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/record", recordHandler)

	fmt.Println("Starting server at port 8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
