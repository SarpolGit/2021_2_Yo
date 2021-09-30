package main

import (
	deliveryAuth "backend/auth/delivery/http"
	localStorageAuth "backend/auth/repository/localstorage"
	useCaseAuth "backend/auth/usecase"
	deliveryEventsManager "backend/eventsManager/delivery/http"
	localStorageEventsManager "backend/eventsManager/repository/localstorage"
	useCaseEventsManager "backend/eventsManager/usecase"
	"bufio"
	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"github.com/swaggo/http-swagger"
	_ "backend/docs"
)

func Preflight(w http.ResponseWriter, r *http.Request) {
	log.Info("In preflight")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS,HEAD")
}

//@title BMSTUSA API
//@version 1.0
//@description TP_2021_GO TEAM YO
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

//@host yobmstu.herokuapp.com
//@BasePath /
//@schemes https

func main() {

	log.Info("Main : start")

	f, err := os.Open("auth/secret.txt")
	if err != nil {
		log.Fatal("Main : can't open file with secret keyword!", err)
	}
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	secret := scanner.Text()
	if err := f.Close(); err != nil {
		log.Fatal("Main : can't close file with secret keyword!", err)
	}

	log.Println(secret)

	port := os.Getenv("PORT")
	if port == "" {
		log.Error("Main : PORT must be set")
		port = "8080"
	}

	r := mux.NewRouter()

	repositoryAuth := localStorageAuth.NewRepositoryUserLocalStorage()
	usecaseAuth := useCaseAuth.NewUseCaseAuth(repositoryAuth, []byte(secret))
	handlerAuth := deliveryAuth.NewHandlerAuth(usecaseAuth)

	repositoryEventsManager := localStorageEventsManager.NewRepositoryEventLocalStorage()
	usecaseEventsManager := useCaseEventsManager.NewUseCaseEvents(repositoryEventsManager)
	handlerEventsManager := deliveryEventsManager.NewHandlerEventsManager(usecaseEventsManager)

	r.HandleFunc("/signup", handlerAuth.SignUp).Methods("POST")
	r.HandleFunc("/login", handlerAuth.SignIn).Methods("POST")
	r.HandleFunc("/user", handlerAuth.User).Methods("GET")
	r.HandleFunc("/events", handlerEventsManager.List)
	r.Methods("OPTIONS").HandlerFunc(Preflight)

	r.PathPrefix("/documentation").Handler(httpSwagger.WrapHandler)

	r.Use(gorilla_handlers.CORS(
		gorilla_handlers.AllowedOrigins([]string{"https://bmstusssa.herokuapp.com"}),
		gorilla_handlers.AllowedHeaders([]string{
			"Accept", "Content-Type", "Content-Length",
			"Accept-Encoding", "X-CSRF-Token", "csrf-token"}),
		gorilla_handlers.AllowCredentials(),
		gorilla_handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
	))

	log.Info("Deploying. Port: ", port)

	errServer := http.ListenAndServe(":"+port, r)
	if errServer != nil {
		log.Error("Main : ListenAndServe error: ", errServer)
	}
}
