package users

import (
	"fmt"
	"time"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers/config"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser - Add a new user to configured gogios databases
func CreateUser(user gogios.User, conf *config.Config) error {
	pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(pass)
	for _, database := range conf.Databases {
		err := database.Database.AddUser(user)
		if err != nil {
			return err
		}
	}

	return nil
}

// Login checks the provided username/password combo against the first configured database
func Login(username, password string, db gogios.Database) map[string]interface{} {
	user, err := db.GetUser(username)
	if err != nil {
		var resp = map[string]interface{}{"status": false, "message": "Username not found"}
		return resp
	}

	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Minute * 100000))

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		var resp = map[string]interface{}{"status": false, "message": "Invalid username and password combo"}
		return resp
	}

	tk := Token{
		UserID: user.ID,
		Name:   user.Name,
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: *&expiresAt,
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		fmt.Println(err)
	}

	var resp = map[string]interface{}{"status": true, "message": "Authentication success"}
	resp["token"] = tokenString
	resp["user"] = user

	return resp
}
