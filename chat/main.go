package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
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
	// Execute: applies a parsed template to the specified data object
	t.templ.Execute(w, r)
}

func main() {
	// command-line arguments
	// address
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse()

	r := newRoom()
	// r.tracer = trace.New(os.Stdout)
	// router
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// goroutine
	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
