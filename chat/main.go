package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

type templateHandler struct {
	filename string
	templ    *template.Template
	// template comilation for only once
	once sync.Once
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		// template.Must: returning (*Template, error) and panics if the error is non-nil.
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}

	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	// Execute: applies a parsed template to the specified data object
	t.templ.Execute(w, data)
}

func main() {
	// loading .env file
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}
	host := os.Getenv("HOST_ADDR")
	port := os.Getenv("HOST_PORT")
	google_client_id := os.Getenv("GOOGLE_AUTH_CLIENT_ID")
	google_secret := os.Getenv("GOOGLE_AUTH_SECRET")

	// command-line arguments
	// address
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse()

	gomniauth.SetSecurityKey(os.Getenv("SECURITY_KEY"))
	gomniauth.WithProviders(
		google.New(google_client_id, google_secret, "http://"+host+":"+port+"/auth/callback/google"),
		github.New(google_client_id, google_secret, "http://"+host+":"+port+"/auth/callback/facebook"),
		facebook.New(google_client_id, google_secret, "http://"+host+":"+port+"/auth/callback/github"),
	)

	r := newRoom()
	// r.tracer = trace.New(os.Stdout)
	// router
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	// goroutine
	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
