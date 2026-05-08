package main

import (
	"html/template"
	"log"
	"net/http"

	"TIENDAPATOS/internal/database"
	"TIENDAPATOS/internal/handlers"
)

func main() {
	// 1. Configuración
	store := database.NewUserStore("api/users.jsonl")

	// Cargamos las plantillas
tmplLogin, _ := template.ParseFiles("ui/templates/login.html")
tmplRegister, _ := template.ParseFiles("ui/templates/register.html")

// ¡NUEVO! Cargamos el perfil
tmplProfile, err := template.ParseFiles("ui/templates/perfil.html")
if err != nil {
    log.Fatalf("error cargando template de perfil: %v", err)
}

// Pasamos TODAS las plantillas al Handler (incluyendo el perfil)
userHandler := handlers.NewUserHandler(tmplLogin, tmplRegister, tmplProfile, store)

	// --- RUTAS --- //

	// Inicio
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmplIndex, _ := template.ParseFiles("ui/templates/index.html")
		tmplIndex.Execute(w, nil)
	})

	// Chaquetas
	http.HandleFunc("/chaquetas", func(w http.ResponseWriter, r *http.Request) {
		tmplChaquetas, _ := template.ParseFiles("ui/templates/chaquetas_nina.html")
		tmplChaquetas.Execute(w, nil)
	})

	// RUTAS DE USUARIO
	http.HandleFunc("/procesar-registro", userHandler.SubmitForm) // Guarda en JSONL
	http.HandleFunc("/procesar-login", userHandler.Login)         // Procesa el inicio de sesión
	http.HandleFunc("/login", userHandler.ShowLogin)
	http.HandleFunc("/registro", userHandler.ShowRegister)
	http.HandleFunc("/logout", userHandler.Logout) // Nueva ruta para cerrar sesión

	// RUTAS PROTEGIDAS (Solo entran si tienen la cookie)
	http.HandleFunc("/perfil", userHandler.AuthMiddleware(userHandler.ShowProfile))
	// http.HandleFunc("/admin", userHandler.AuthMiddleware(userHandler.ShowAdmin))


		// Usando la raíz del proyecto directamente 
	fs := http.FileServer(http.Dir("ui/static")) // Sin el ./  ni / al final
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Servidor escuchando en http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//go run ./cmd/web/main.go
