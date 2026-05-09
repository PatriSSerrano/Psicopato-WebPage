package handlers

import (
	"html/template"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"

	"TIENDAPATOS/internal/database"
	"TIENDAPATOS/internal/models"
)

type UserHandler struct {
	templates *template.Template // Cargaremos todos los archivos aquí
	store *database.UserStore
}

func NewUserHandler(tmpl *template.Template, store *database.UserStore) *UserHandler {
	return &UserHandler{
		templates: tmpl,
		store: store,
	}
}

//Métodos de visualización
func (h *UserHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
    h.templates.ExecuteTemplate(w, "login.html", nil)
}

// Método para mostrar el perfil dinámico
func (h *UserHandler) ShowProfile(w http.ResponseWriter, r *http.Request) {
    // 1. Leemos la cookie para saber quién es
    cookie, err := r.Cookie("session_user")
    if err != nil {
        http.Redirect(w, r, "/login?error=no_autorizado", http.StatusSeeOther)
        return
    }
    
    // 2. Buscamos al usuario en la base de datos (tu JSON)
    email := cookie.Value
    user, err := h.store.GetUserByEmail(email)
    if err != nil {
        http.Redirect(w, r, "/login?error=no_autorizado", http.StatusSeeOther)
        return
    }

    // 3. Empaquetamos los datos dinámicos
    data := map[string]interface{}{
        "Nombre": user.Name,
        "Email":  user.Email,
    }

    // 4. Inyectamos los datos en la plantilla perfil.html
    // Fíjate que aquí NO pasamos nil, pasamos 'data'
    h.templates.ExecuteTemplate(w, "perfil.html", data)
}

func (h *UserHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
    h.templates.ExecuteTemplate(w, "register.html", nil)
}

// SubmitForm procesa el registro de nuevos usuarios
func (h *UserHandler) SubmitForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()

	password := r.FormValue("password") // Cogemos la contraseña del HTML

	//CIFRAR LA CONTRASEÑA 
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	//Crear el usuario con el Hash
	user := models.User{
		Name:         r.FormValue("nombre"),
		Email:        r.FormValue("email"),
		PasswordHash: string(hash), // Guardamos el resumen
	}

	// Si hay error al guardar:
	if err := h.store.AppendUser(user); err != nil {
		// Redirigimos al registro con un aviso de error
		http.Redirect(w, r, "/registro?error=servidor", http.StatusSeeOther)
		return
	}
	// SI TODO VA BIEN: Redirigimos al login con mensaje de éxito
	http.Redirect(w, r, "/login?exito=registro", http.StatusSeeOther)

}
// Login procesa el intento de entrada
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()

	email := r.FormValue("email")
	password := r.FormValue("password")

	// 1. Buscar al usuario en la base de datos
	// Si el usuario no existe o la contraseña está mal:
	user, err := h.store.GetUserByEmail(email)
	if err != nil {
		http.Redirect(w, r, "/login?error=credenciales", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login?error=credenciales", http.StatusSeeOther)
		return
	}

		// Creamos la cookie de sesión
	cookie := &http.Cookie{
		Name:     "session_user",
		Value:    email, // Guardamos el email como identificador
		Path:     "/",
		HttpOnly: true, // Seguridad: impide que JavaScript acceda a la cookie
		MaxAge:   3600, // Dura 1 hora
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	log.Printf("INFO: Cookie de sesión creada para %s", email)

	http.Redirect(w, r, "/?exito=login", http.StatusSeeOther)


	// SI EL LOGIN ES CORRECTO: Redirigimos a la página principal
	http.Redirect(w, r, "/?exito=login", http.StatusSeeOther)
}

// Middleware para proteger rutas
func (h *UserHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_user")
        if err != nil {
            // Si no hay cookie, lo mandamos al login
            log.Printf("WARN: Intento de acceso no autorizado a %s", r.URL.Path)
            http.Redirect(w, r, "/login?error=no_autorizado", http.StatusSeeOther)
            return
        }
        
        // Si hay cookie, dejamos que pase a la siguiente función (next)
        log.Printf("INFO: Usuario %s accediendo a ruta protegida", cookie.Value)
        next.ServeHTTP(w, r)
    }
}

// Logout cierra la sesión del usuario eliminando la cookie
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
    // Verificamos que sea un método POST por seguridad
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    // Sobrescribimos la cookie actual con una caducada
    cookie := &http.Cookie{
        Name:     "session_user",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        MaxAge:   -1, // Obliga al navegador a borrarla al instante
        Expires:  time.Now().Add(-1 * time.Hour), // Fecha en el pasado por si acaso
    }
    http.SetCookie(w, cookie)

    log.Println("INFO: Un usuario ha cerrado sesión correctamente.")

    // Redirigimos a la página de login con un mensaje de éxito
    http.Redirect(w, r, "/login?exito=logout", http.StatusSeeOther)
}

func (h *UserHandler) ShowHome(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_user")
    isLoggedIn := (err == nil)

    data := map[string]interface{}{
        "Titulo":     "Inicio - Tienda de Patos",
        "Logueado":   isLoggedIn,
        "UserEmail":  "",
    }
    
    if isLoggedIn {
        data["UserEmail"] = cookie.Value
    }

    // Usamos ExecuteTemplate especificando el archivo
    h.templates.ExecuteTemplate(w, "index.html", data)
}