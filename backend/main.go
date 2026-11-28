package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)


type User struct {
	Id 		int 	`json:"id"`
	Name	string	`json:"name"`
	Email	string	`json:"email"`
}

// main function

func main() {
	// database connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"),
)
	// IF ERROR	 log the error and close the connection to DB
	if err != nil{
		log.Fatal(err)
	}

	defer db.Close()


	//Create table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
	
	if err != nil {
		log.Fatal(err)
	}


	// create router
	router := mux.NewRouter()
	router.HandleFunc("/api/go/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/api/go/create/user", createUser(db)).Methods("POST")
	router.HandleFunc("/api/go/user/{id}", getUser(db)).Methods("GET")
	router.HandleFunc("/api/go/user/update/{id}", updateUser(db)).Methods("PUT")
	router.HandleFunc("/api/go/delete/user", deleteUser(db)).Methods("DELETE")

	enhanceRouter := enableCORS(jsonContentTypeMiddleware(router))

	log.Fatal(http.ListenAndServe(":8000", enhanceRouter))

}


func enableCORS(next http.Handler)http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		//check if the request is for CORS preflight


		if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK) 
		return
		}

		next.ServeHTTP(w, r)
	})
}


func jsonContentTypeMiddleware(next http.Handler)http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getUsers (db *sql.DB)http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		
		users := []User{} //array of users
		for rows.Next() {
			var u User 
			if err := rows.Scan(&u.Id, &u.Name, &u.Email);err != nil {
				log.Fatal(err)
			}
			users = append(users, u)
		}
		if errr := rows.Err(); err != nil {
			log.Fatal(errr)
		}
		
		json.NewEncoder(w).Encode(users)
	}
}

func getUser(db *sql.DB)http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r) 
		id := vars["id"]

		var u User

		err := db.QueryRow("SELECT * FROM users WHERE id - $1", id).Scan(&u.Id, &u.Name, &u.Email)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return 
		}

		json.NewEncoder(w).Encode(u)

	}
}


// crete user 
func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		json.NewDecoder(r.Body).Decode(&u)

		err := db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING ID", u.Name, u.Email).Scan(&u.Id)

		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(u)
	}
}


// update user

func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter,r *http.Request) {
		var u User
		json.NewDecoder(r.Body).Decode(&u)

		vars := mux.Vars(r)
		id:= vars["id"]

		_, err := db.Exec("update users set name = $1, email =  $2 where id = $3", u.Name, u.Email, id)

		if err != nil {
			log.Fatal(err)
		}


		// Retrieve the updated user data from the database
		var updatedUser User 
		err = db.QueryRow("SELECT id, name, email FROM users where id = $1 ", id).Scan(&updatedUser.Id, &updatedUser.Name, &updatedUser.Email)
		if err != nil {
			log.Fatal(err)
		}


		// send the updated user data in the response
		json.NewEncoder(w).Encode(updatedUser)
	}
}

// delete user
func deleteUser (db *sql.DB) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id:= vars["id"]

		var u User
		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&u.Id, &u.Name, &u.Email)
		if err !=  nil {
			w.WriteHeader(http.StatusNotFound)
			return 
		}else {
			_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
			if err != nil {
				// todo : fix error handling
				w.WriteHeader(http.StatusNotFound)
				return  
			}

			json.NewEncoder(w).Encode("user deleted")
		}
	}

}