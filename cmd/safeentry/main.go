package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/kvclient"
	"log"
	"net/http"
	"time"
)

type HistoryRecord struct {
	Location string
	CheckIn  bool
}

type GetHistoryRequest struct {
	IC       string
	Phone    string
}

type CheckInRequest struct {
	IC       string
	Phone    string
	Location string
	CheckIn  bool
}

func main() {
	// Read parameters
	address := flag.String("address", ":8080", "server port")
	configPath := flag.String("config", "./configs/default_client.json", "config path")
	flag.Parse()

	// Connect to DB
	config, err := kvclient.NewClientConfigFromFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	db := kvclient.NewKeyValueStoreClient(config)

	// HTTP Handlers
	fs := http.FileServer(http.Dir("./web/safe-entry/build"))
	http.Handle("/", fs)
	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		// create context
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// parse
		var req GetHistoryRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Println("Error when parse request:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		key := []byte(fmt.Sprintf("safeentry:ic:%v:phone:%v", req.IC, req.Phone))

		// get previous history
		history := make([]*HistoryRecord, 0)

		getRes, err := db.Get(ctx, &pb.GetRequest{Key: key})
		if err == nil && getRes.FoundKey == pb.FoundKey_KEY_FOUND {
			historyBytes := getRes.Object
			buf := bytes.NewBuffer(historyBytes)
			dec := gob.NewDecoder(buf)
			_ = dec.Decode(&history)
		}

		// return history
		historyJson, err := json.Marshal(history)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, string(historyJson))
	})

	http.HandleFunc("/api/checkin", func(w http.ResponseWriter, r *http.Request) {
		// create context
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// parse
		var req CheckInRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Println("Error when parse request:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		key := []byte(fmt.Sprintf("safeentry:ic:%v:phone:%v", req.IC, req.Phone))

		// get previous history
		history := make([]*HistoryRecord, 0)

		getRes, err := db.Get(ctx, &pb.GetRequest{Key: key})
		if err == nil && getRes.FoundKey == pb.FoundKey_KEY_FOUND {
			historyBytes := getRes.Object
			buf := bytes.NewBuffer(historyBytes)
			dec := gob.NewDecoder(buf)
			_ = dec.Decode(&history)
		}

		// append new record
		history = append(history, &HistoryRecord{
			Location: req.Location,
			CheckIn:  req.CheckIn,
		})

		// put new history to db
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err = enc.Encode(history)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = db.Put(ctx, &pb.PutRequest{
			Key:    key,
			Object: buf.Bytes(),
		})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// return new history
		historyJson, err := json.Marshal(history)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, string(historyJson))
	})

	log.Printf("Listening on %v", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
