package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AppData maneja el JSON dinámico que viene del frontend
type AppData map[string]interface{}

var collection *mongo.Collection

func main() {
	// Preparado para entorno local o variables de entorno (ej. en Render)
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		//mongoURI = "mongodb://localhost:27017"
		mongoURI = "mongodb+srv://mauro:mrmtools..@cluster0.buufcuz.mongodb.net/?appName=Cluster0"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Error conectando a Mongo:", err)
	}

	// Usaremos una base de datos "mrmtools" y la colección "expensescalendar"
	collection = client.Database("mrmtools").Collection("expensescalendar")

	// Rutas
	http.HandleFunc("/api/sync", corsMiddleware(syncHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor corriendo en el puerto %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// corsMiddleware permite que tu frontend en HTML puro se conecte sin bloqueos de CORS
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Identificador único del usuario (por ahora quemado como "user_data")
	filter := bson.M{"_id": "user_data"}

	switch r.Method {
		case http.MethodGet:
			// Consultar datos
			var result AppData
			err := collection.FindOne(ctx, filter).Decode(&result)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					// Si no hay documentos, devolvemos un 204 No Content
					w.WriteHeader(http.StatusNoContent)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result["payload"])

		case http.MethodPost:
			// Guardar datos
			var payload interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON Inválido", http.StatusBadRequest)
				return
			}

			// Upsert: Si existe lo actualiza, si no, lo crea
			update := bson.M{"$set": bson.M{"payload": payload}}
			opts := options.Update().SetUpsert(true)

			_, err := collection.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"success"}`))
			
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}