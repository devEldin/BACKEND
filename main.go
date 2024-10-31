package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Zadatak struct {
	ID      string `json:"id,omitempty" bson:"_id,omitempty"`
	Zadatak string `json:"zadatak" bson:"zadatak"`
}

var klijent *mongo.Client

func poveziMongoDB() {
	var err error
	connectionString := "mongodb+srv://eldinspj:001122@cluster0.ztbhl.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	klijent, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = klijent.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Nije se moguce povezati sa MongoDB :", err)
	}
	fmt.Println("Uspjesno povezano sa MongoDB!")
}

func dodajZadatakHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		var zadatak Zadatak
		err := json.NewDecoder(r.Body).Decode(&zadatak)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		kolekcija := klijent.Database("eldin").Collection("test")
		_, err = kolekcija.InsertOne(context.TODO(), zadatak)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Primljen zadatak: %s", zadatak.Zadatak)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Zadatak uspješno dodan"})
	}
}

func dohvatiZadatkeHandler(w http.ResponseWriter, r *http.Request) {
	// log.Println("Primljen zahtjev na /dohvatiZadatkeHandler")

	// CORS za komunikacioju izmedju baackenda i frontenda
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Provjeraa sa serverom da li je request safe
	if r.Method == http.MethodOptions {
		log.Println("Obrada OPTIONS zahtjeva za CORS preflight")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Obradi stvarni GET zahtjev
	if r.Method == http.MethodGet {
		log.Println("Obrada GET zahtjeva")

		kolekcija := klijent.Database("eldin").Collection("test")

		// Pokušaj pronaći dokumente u kolekciji sa praznim filterom
		cursor, err := kolekcija.Find(context.TODO(), bson.M{})
		if err != nil {
			log.Printf("Greška pri pronalaženju dokumenata u kolekciji: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.TODO())

		var zadaci []bson.M
		// log.Println("Prolazak kroz cursor za dekodiranje zadataka")

		for cursor.Next(context.TODO()) {
			var zadatak bson.M
			if err := cursor.Decode(&zadatak); err != nil {
				log.Printf("Greška pri dekodiranju zadatka iz cursora: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			zadaci = append(zadaci, zadatak)
			log.Println("Zadatak dekodiran i dodan:", zadatak)
		}

		// Provjeri grešku nakon iteracije cursora
		if err := cursor.Err(); err != nil {
			log.Printf("Cursor je naišao na grešku: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// log.Println("Svi zadaci su dohvaćeni, šaljem odgovor")
		json.NewEncoder(w).Encode(zadaci) // Kodiraj i pošalji sve zadatke kao JSON
	} else {
		// Ako metoda nije GET ili OPTIONS, vrati grešku Method Not Allowed
		log.Printf("Metoda %s nije dozvoljena, vraćam 405\n", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	poveziMongoDB()
	defer func() {
		if err := klijent.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	http.HandleFunc("/api/tasks", dodajZadatakHandler)
	http.HandleFunc("/api/tasks/get", dohvatiZadatkeHandler)
	log.Println("Server pokrenut na portu 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
