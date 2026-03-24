package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	// 1. Definimos el Context Path y el Endpoint
	contextPath := "/mytestapp"
	apiPath := "/api/movements/new"
	fullPath := contextPath + apiPath

	// 2. Creamos el enrutador
	mux := http.NewServeMux()

	// 3. Registramos el endpoint específico
	mux.HandleFunc(fullPath, func(w http.ResponseWriter, r *http.Request) {
		// Solo aceptamos POST (típico para 'new' movements)
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"message": "Movement created successfully", "status": "ok"}`)
		} else {
			// Si no es POST, devolvemos error
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 4. Configuración del puerto para Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Servidor escuchando en: http://localhost:%s%s\n", port, fullPath)

	// 5. Arrancamos el servidor usando nuestro 'mux'
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		panic(err)
	}
}