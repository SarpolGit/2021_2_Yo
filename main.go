package main

import (
	deliveryAuth "backend/auth/delivery/http"
	localStorageAuth "backend/auth/repository/localstorage"
	useCaseAuth "backend/auth/usecase"
	//"github.com/rs/cors"
	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

//Установка допустимых параметров запросов с фронта (или наоборот - на фронт, не оч понял)
func Preflight(w http.ResponseWriter, r *http.Request) {
	log.Info("In preflight")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS,HEAD")
}

/*
Курлы:

Регистрация:
curl -X POST -H "Content-Type: application/json" -d '{"name": "John", "surname": "Brown", "email": "lol@mail.ru", "password": "1224"}' http://localhost:8080/signup

Логин:

Список пользователей:
~(master*) » curl -X GET -H "Content-Type: application/json" -d '' http://localhost:8080/list

User:




 */

func main() {
	log.Println("Hello, World!")

	port := os.Getenv("PORT")
	if port == "" {
		log.Error("$PORT must be set")
	}

	r := mux.NewRouter()

	repo := localStorageAuth.NewRepositoryUserLocalStorage()
	useCase := useCaseAuth.NewUseCaseAuth(repo)
	handler := deliveryAuth.NewHandlerAuth(useCase)
	//r.Use(handler.MiddleWare)

	r.HandleFunc("/signup", handler.SignUp).Methods("POST")
	r.HandleFunc("/signin", handler.SignIn).Methods("POST")
	r.HandleFunc("/list", handler.List).Methods("GET")
	r.HandleFunc("/user", handler.User).Methods("GET")
	r.Methods("OPTIONS").HandlerFunc(Preflight)
	//Нужен метод для SignIn с методом GET

	r.Use(gorilla_handlers.CORS(
		gorilla_handlers.AllowedOrigins([]string{"https://bmstusssa.herokuapp.com"}),
		gorilla_handlers.AllowedHeaders([]string{
			"Accept", "Content-Type", "Content-Length",
			"Accept-Encoding", "X-CSRF-Token", "csrf-token", "Authorization"}),
		gorilla_handlers.AllowCredentials(),
		gorilla_handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
	))

	log.Info("Deploying. Port: ", port)

	//err := http.ListenAndServe(":"+port, r)
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Error("main error: ", err)
	}

}
