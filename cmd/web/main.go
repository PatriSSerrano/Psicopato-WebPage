package main

import (
	"html/template"
	"log"
	"net/http"

	"TIENDAPATOS/internal/database"
	"TIENDAPATOS/internal/handlers"
)

func main() {
	//Configuración
	store := database.NewUserStore("api/users.jsonl")

	// Cargamos todos los .html de la carpeta templates
	// El patrón "ui/templates/*.html" agarra todos los archivos automáticamente
	tmpl, err := template.ParseGlob("ui/templates/*.html")
	if err != nil {
		log.Fatalf("Error cargando templates: %v", err)
	}
	// inicializamos Handler
	userHandler := handlers.NewUserHandler(tmpl, store)

	// --- RUTAS --- //

	// Inicio
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	if r.URL.Path != "/" {
	// 		http.NotFound(w, r)
	// 		return
	// 	}
	// 	tmplIndex, _ := template.ParseFiles("ui/templates/index.html")
	// 	tmplIndex.Execute(w, nil)
	// })

	// Chaquetas
	http.HandleFunc("/chaquetas", func(w http.ResponseWriter, r *http.Request) {
		tmplChaquetas, _ := template.ParseFiles("ui/templates/chaquetas_nina.html")
		tmplChaquetas.Execute(w, nil)
	})

	// 3. RUTAS (Asegúrate de que la raíz "/" apunte a un handler, no a un FileServer)
	http.HandleFunc("/", userHandler.ShowHome)
	http.HandleFunc("/login", userHandler.ShowLogin)
	http.HandleFunc("/registro", userHandler.ShowRegister)
	http.HandleFunc("/perfil", userHandler.AuthMiddleware(userHandler.ShowProfile))
	// http.HandleFunc("/admin", userHandler.AuthMiddleware(userHandler.ShowAdmin))
	// RUTAS DE USUARIO
	http.HandleFunc("/procesar-registro", userHandler.SubmitForm) // Guarda en JSONL
	http.HandleFunc("/procesar-login", userHandler.Login)         // Procesa el inicio de sesión
	http.HandleFunc("/logout", userHandler.Logout) // Nueva ruta para cerrar sesión


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
