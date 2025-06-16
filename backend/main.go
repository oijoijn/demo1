package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

/* ---------- モデル ---------- */

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *sql.DB

/* ---------- エントリポイント ---------- */

func main() {
	/* --- DB 接続 & リトライ --- */
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(mysql:3306)/my_app_db?parseTime=true"
	}

	var err error
	var retry int
	for {
		db, err = sql.Open("mysql", dsn)
		if err == nil && db.Ping() == nil {
			break // 接続成功
		}
		log.Printf("MySQL 接続失敗 (再試行 %d): %v\n", retry+1, err)
		retry++
		if retry > 15 {
			log.Fatal("DB接続エラー: 試行回数の上限に達しました")
		}
		time.Sleep(2 * time.Second)
	}
	defer db.Close()
	log.Println("✅ データベースに接続しました")

	/* --- ルーター --- */
	r := mux.NewRouter()

	// ヘルスチェック
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	// CRUD エンドポイント
	r.HandleFunc("/api/users", getUsers).Methods(http.MethodGet)
	r.HandleFunc("/api/users", createUser).Methods(http.MethodPost)
	r.HandleFunc("/api/users/{id}", updateUser).Methods(http.MethodPut, http.MethodOptions)
	r.HandleFunc("/api/users/{id}", deleteUser).Methods(http.MethodDelete, http.MethodOptions)

	// CORS ミドルウェア
	handler := corsMiddleware(r)

	/* --- サーバー起動 --- */
	log.Println("🚀 サーバーをポート8080で起動します...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

/* ---------- ミドルウェア ---------- */

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

/* ---------- ハンドラ ---------- */

// 全ユーザー取得 (Read)
func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, email FROM users ORDER BY id DESC")
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
	if len(users) == 0 {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(users)
}

// ユーザー作成 (Create)
func createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", u.Name, u.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	u.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

// ユーザー更新 (Update)
func updateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", u.Name, u.Email, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// ユーザー削除 (Delete)
func deleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if _, err := db.Exec("DELETE FROM users WHERE id = ?", id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
