package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   string `json:"age"`
}

func main() {
	// Database connection
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/go_crud")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Define routes
	http.HandleFunc("GET /users", getUsers(db))
	http.HandleFunc("POST /create-users", createUser(db))
	http.HandleFunc("PUT /update-user", updateUser(db))

	// Start the server
	fmt.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// read user
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, email FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		// Get name and email from form data
		name := r.FormValue("name")
		email := r.FormValue("email")
		age := r.FormValue("age")

		// Validate input
		if name == "" || email == "" || age == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		// Insert user into database
		result, err := db.Exec("INSERT INTO users (name, email, age) VALUES (?, ?, ?)", name, email, age)
		if err != nil {
			http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the ID of the newly inserted user
		id, _ := result.LastInsertId()

		// Create user object
		user := User{
			ID:    int(id),
			Name:  name,
			Email: email,
			Age:   age,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		// Get user ID and other fields from form data
		idStr := r.FormValue("id")
		name := r.FormValue("name")
		email := r.FormValue("email")
		age := r.FormValue("age")

		// Convert ID to int
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Validate input
		if name == "" && email == "" && age == "" {
			http.Error(w, "At least one field (name, email, or age) is required for update", http.StatusBadRequest)
			return
		}

		// Prepare SQL query
		query := "UPDATE users SET"
		params := []interface{}{}
		if name != "" {
			query += " name = ?,"
			params = append(params, name)
		}
		if email != "" {
			query += " email = ?,"
			params = append(params, email)
		}
		if age != "" {
			query += " age = ?,"
			params = append(params, age)
		}
		query = query[:len(query)-1] // Remove trailing comma
		query += " WHERE id = ?"
		params = append(params, id)

		// Execute update query
		result, err := db.Exec(query, params...)
		if err != nil {
			http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "User not found or no changes made", http.StatusNotFound)
			return
		}

		// Fetch updated user
		var user User
		err = db.QueryRow("SELECT id, name, email, age FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name, &user.Email, &user.Age)
		if err != nil {
			http.Error(w, "Failed to fetch updated user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}
