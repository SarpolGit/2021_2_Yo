package http

import (
	"backend/auth"
	"backend/models"
	"fmt"
	//"backend/models"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var cookies = make(map[string]string)

const (
	STATUS_OK    = "ok"
	STATUS_ERROR = "error"
)

type HandlerAuth struct {
	useCase auth.UseCase
}

func NewHandlerAuth(useCase auth.UseCase) *HandlerAuth {
	//auth.UseCase - это чистый интерфейс
	//Передаём интерфейс, а не конкретную реализацию, поскольку нужно будет передавать мок для тестирования
	return &HandlerAuth{
		useCase: useCase,
	}
}

//Структура, в которую мы попытаемся перевести JSON-запрос
//Эта структура - неполная, она, например, не содержит ID и чего-нибудь ещё (дату рождения, например)
type userDataForSignUp struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Mail     string `json:"email"`
	Password string `json:"password"`
}

type userDataForSignIn struct {
	Mail     string `json:"email"`
	Password string `json:"password"`
}

type userDataForResponse struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Mail    string `json:"mail"`
}

func makeUserDataForResponse(user *models.User) *userDataForResponse {
	return &userDataForResponse{
		Name:    user.Name,
		Surname: user.Surname,
		Mail:    user.Mail,
	}
}

func getUserFromJSONSignUp(r *http.Request) (*userDataForSignUp, error) {
	userInput := new(userDataForSignUp)
	//Пытаемся декодировать JSON-запрос в структуру
	//Валидность данных проверяется на фронтенде (верно?...)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		return nil, err
	}
	return userInput, nil
}

func getUserFromJSONSignIn(r *http.Request) (*userDataForSignIn, error) {
	userInput := new(userDataForSignIn)
	//Пытаемся декодировать JSON-запрос в структуру
	//Валидность данных проверяется на фронтенде (верно?...)
	err := json.NewDecoder(r.Body).Decode(userInput)
	if err != nil {
		return nil, err
	}
	return userInput, nil
}

//Не уверен, что здесь указатель, проверить!
func (h *HandlerAuth) setCookieWithJwtToken(w http.ResponseWriter, userMail, userPassword string) {
	/////////
	log.Info("setCookieWithJwtToken : started")
	/////////
	//TODO: Сделать так, чтобы SignIn возвращал только токен и ошибку. информация о user будет возвращаться в User (функция)
	//TODO: Вроде сделал
	jwtToken, err := h.useCase.SignIn(userMail, userPassword)
	/////////
	log.Info("setCookieWithJwtToken : jwtToken = ", jwtToken)
	/////////
	if err == auth.ErrUserNotFound {
		/////////
		log.Error("SignIn : setCookieWithJwtToken error")
		/////////
		http.Error(w, `{"error":"signin_user_not_found"}`, 500)
		return
	}
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   true,
	}
	//Костыль, добавляем ещё одну куку, которая не записывается голангом
	http.SetCookie(w, cookie)
	cs := (w).Header().Get("Set-Cookie")
	cs += "; SameSite=None"
	(w).Header().Set("Set-Cookie", cs)
	/////////
	log.Info("setCookieWithJwtToken : ended")
	/////////
}

func (h *HandlerAuth) SignUp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	/////////
	log.Info("SignUp : started")
	/////////
	userFromRequest, err := getUserFromJSONSignUp(r)
	/////////
	log.Info("SignUp : userFromRequest = ", userFromRequest)
	/////////
	if err != nil {
		/////////
		log.Error("SignUp : didn't get user from JSON")
		/////////
		http.Error(w, `{"error":"signup_json"}`, 500)
		return
	}
	err = h.useCase.SignUp(userFromRequest.Name, userFromRequest.Surname, userFromRequest.Mail, userFromRequest.Password)
	if err != nil {
		/////////
		log.Error("SignUp : SignUp error")
		/////////
		http.Error(w, `{"error":"signup_signup"}`, 500)
		return
	}
	//TODO: Поставить Cookie с jwt-токеном при регистрации
	//TODO: Вроде сделал
	h.setCookieWithJwtToken(w, userFromRequest.Mail, userFromRequest.Password)
	/////////
	log.Info("SignUp : ended")
	/////////
	return
}

func (h *HandlerAuth) SignIn(w http.ResponseWriter, r *http.Request) {
	/////////
	log.Info("SignIn : started")
	/////////
	defer r.Body.Close()
	userFromRequest, err := getUserFromJSONSignIn(r)
	/////////
	log.Info("SignIn : userFromRequest = ", userFromRequest)
	/////////
	if err != nil {
		/////////
		log.Error("SignIn : getUserFromJSON error")
		/////////
		http.Error(w, `{"error":"signin_json"}`, 500)
		return
	}
	h.setCookieWithJwtToken(w, userFromRequest.Mail, userFromRequest.Password)
	/////////
	log.Info("SignIn : ended")
	/////////
	return
}

func (h *HandlerAuth) List(w http.ResponseWriter, r *http.Request) {
	/////////
	log.Info("List : started")
	/////////
	fmt.Println("")
	fmt.Println("=============================")
	fmt.Println("=========U==S==E==R==S=======")
	fmt.Println("=============================")
	defer r.Body.Close()
	users := h.useCase.List()
	for _, user := range users {
		fmt.Println(user)
		userData := makeUserDataForResponse(&user)
		userDataToWrite, _ := json.Marshal(userData)
		w.Write(userDataToWrite)
	}
	fmt.Println("=============================")
	fmt.Println("=-=-=-=-=-=-=-=-=-=-=-=-=-=-=")
	fmt.Println("=============================")
	fmt.Println("")
	/////////
	log.Info("List : ended")
	/////////
	return
}

func (h *HandlerAuth) MiddleWare(handler http.Handler) http.Handler {
	/////////
	log.Info("MiddleWare : started & ended")
	/////////
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS,HEAD")
		handler.ServeHTTP(w, r)
	})
}

func (h *HandlerAuth) User(w http.ResponseWriter, r *http.Request) {
	/////////
	log.Info("User : started")
	/////////
	//TODO: Отправить информацию о пользователе таким образом:
	/*
		//Получаем данные о пользователе для того, чтобы отправить их пользователю
		userData := makeUserDataForResponse(foundUser)
		w.WriteHeader(http.StatusOK)
		userDataToWrite, err := json.Marshal(userData)
		if err != nil {
			/////////
			log.Error("User : json.Marshall error")
			/////////
			return
		}
		w.Write(userDataToWrite)
	 */

	defer r.Body.Close()
	cookie, err := r.Cookie("session_id")
	/////////
	log.Info("User : cookie.value = ", cookie.Value)
	/////////
	if err != nil {
		/////////
		log.Error("User : getting cookie error")
		/////////
		w.WriteHeader(http.StatusTeapot)
		return
	}
	//TODO: Разобраться, как работает ParseToken и что возвращает
	userID, err := h.useCase.ParseToken(cookie.Value)
	/////////
	log.Info("User : userID = ", userID)
	/////////
	if err != nil {
		/////////
		log.Info("User : parse error")
		/////////
		w.WriteHeader(http.StatusTeapot)
		return
	}
	w.WriteHeader(http.StatusOK)
	//TODO: отправить информацию пользователю
	response := makeUserDataForResponse(&models.User{
		ID:       userID,
		Name:     "Faked name",
		Surname:  "Faked surname",
		Mail:     "FakedMail@mail.ru",
		Password: "",
	})
	b, err := json.Marshal(response)
	if err != nil {
		/////////
		log.Info("User : marshal error")
		/////////
	}
	w.Write(b)
	/////////
	log.Info("User : ended")
	/////////
}
