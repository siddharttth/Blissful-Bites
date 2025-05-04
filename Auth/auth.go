package Auth

import (
	"blissfulbites/DB"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	Login(username, password string) bool
	Signup(username, password string) bool
}

type JSONAuth struct {
	users []User
	mu    sync.Mutex
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewJSONAuth() *JSONAuth {
	ja := &JSONAuth{}
	err := ja.LoadUsers()
	if err != nil {
		fmt.Println("Error loading users:", err)
	}
	return ja
}

func (ja *JSONAuth) LoadUsers() error {
	data, err := ioutil.ReadFile("users.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &ja.users)
	if err != nil {
		return err
	}
	return nil
}

func (ja *JSONAuth) SaveUsers() error {
	data, err := json.Marshal(ja.users)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("users.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (ja *JSONAuth) Login(username, password string) bool {
	ja.mu.Lock()
	defer ja.mu.Unlock()
	for _, user := range ja.users {
		if user.Username == username {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			return err == nil
		}
	}
	return false
}

func (ja *JSONAuth) Signup(username, password string) bool {
	ja.mu.Lock()
	defer ja.mu.Unlock()
	for _, user := range ja.users {
		if user.Username == username {
			return false
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return false
	}
	ja.users = append(ja.users, User{Username: username, Password: string(hashedPassword)})
	if err := ja.SaveUsers(); err != nil {
		fmt.Println("Error saving user:", err)
		return false
	}
	return true
}

// DBAuth implements Auth interface using PostgreSQL DB
type DBAuth struct {
	mu sync.Mutex
}

func NewDBAuth() *DBAuth {
	return &DBAuth{}
}

func (dba *DBAuth) Signup(username, password string) bool {
	dba.mu.Lock()
	defer dba.mu.Unlock()

	// Check if email already exists
	exists, err := DB.CheckEmailExists(username)
	if err != nil {
		fmt.Println("Error checking email existence:", err)
		return false
	}
	if exists {
		return false
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return false
	}

	// Insert user credentials into user_credentials table
	query := `
		INSERT INTO user_credentials (email, password)
		VALUES ($1, $2)
	`

	_, err = DB.DB.Exec(query, username, string(hashedPassword))
	if err != nil {
		fmt.Println("Error inserting user credentials:", err)
		return false
	}

	return true
}

func (dba *DBAuth) Login(username, password string) bool {
	dba.mu.Lock()
	defer dba.mu.Unlock()

	// Query password hash from user_credentials table
	var hashedPassword string
	query := "SELECT password FROM user_credentials WHERE email = $1"
	err := DB.DB.QueryRow(query, username).Scan(&hashedPassword)
	if err != nil {
		fmt.Println("Error querying user credentials:", err)
		return false
	}

	// Compare password hash
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// Get Google OAuth credentials from environment variables
func GetGoogleOAuthCredentials() (clientID, clientSecret string) {
	clientID = os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	return
}
