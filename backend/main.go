package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time" // timeパッケージをインポート

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// User構造体
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *sql.DB

// メイン関数
func main() {
	// データベース接続
	var err error
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:3306)/my_app_db?parseTime=true"
	}

	// DB接続リトライ処理
	var counts int64
	for {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Println("MySQLへの接続設定に失敗しました:", err)
			time.Sleep(2 * time.Second)
			counts++
			if counts > 15 {
				log.Fatal("DB接続エラー: 試行回数の上限に達しました。")
			}
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Println("MySQLへのPingに失敗しました (再試行します):", err)
			time.Sleep(2 * time.Second)
			counts++
			if counts > 15 {
				log.Fatal("DB接続エラー: 試行回数の上限に達しました。")
			}
			continue
		}
		break
	}

	defer db.Close()
	log.Println("データベースに接続しました")

	// ルーターの初期化
	r := mux.NewRouter()

	// APIエンドポイントの定義
	r.HandleFunc("/api/users", getUsers).Methods("GET")
	r.HandleFunc("/api/users", createUser).Methods("POST")
	r.HandleFunc("/api/users/{id}", updateUser).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/users/{id}", deleteUser).Methods("DELETE", "OPTIONS")

	// CORSミドルウェアを適用
	handler := corsMiddleware(r)

	// サーバー起動
	log.Println("サーバーをポート8080で起動します...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// --- CORSミドルウェア ---
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- ハンドラ関数 ---

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
	if users == nil {
		// ユーザーが一人もいない場合は空のJSON配列を返す
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

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

// ユーザー更新 (Update)
func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

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
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
