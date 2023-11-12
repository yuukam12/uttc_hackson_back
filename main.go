package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/oklog/ulid"
)

type ContentResForHTTPGet struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Url         string    `json:"url"`
	Image       []byte    `json:"image"`
	UploadedBy  string    `json:"uploaded_by"`
	CreateDate  time.Time `json:"create_date"`
	Category    string    `json:"category"`
	Media       string    `json:"media"`
}

type ContentReqForHTTPPost struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url"`
	Image       []byte `json:"image,omitempty"`
	UploadedBy  string `json:"uploaded_by"`
	Category    string `json:"category"`
	Media       string `json:"media"`
}

// ① GoプログラムからMySQLへ接続
var db *sql.DB

func init() {
	// ①-1: 環境変数からMySQL接続情報を取得
	//envFilePath := "./.env_mysql"
	//if err := godotenv.Load(envFilePath); err != nil {
	//	log.Fatalf("fail: godotenv.Load, %v\n", err)
	//}
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	//connStr := fmt.Sprintf("%s:%s@(localhost:3306)/%s", mysqlUser, mysqlPwd, mysqlDatabase)
	// _db, nil := sql.Open("mysql", connStr)
	_db, nil := sql.Open("mysql", connStr)

	// ①-3: データベースへのPingを確認
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}
	db = _db
}

// ② /contentでリクエストされた時の処理
func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "https://uttc-hackson-front-pyjlg4ck5-yuukmatsumotos-projects.vercel.app")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	switch r.Method {
	case http.MethodOptions:
		log.Printf("options")
		w.WriteHeader(http.StatusOK)
		return

	case http.MethodGet:
		// GETリクエストの処理
		category := r.URL.Query().Get("category")
		keyword := r.URL.Query().Get("keyword")

		if category == "" {
			log.Println("fail: Category is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		query := "SELECT id, title, description, url, image, uploaded_by, category, media FROM content WHERE category = ?"
		if keyword != "" {
			query += " AND (title LIKE ? OR description LIKE ?)"
		}

		var rows *sql.Rows
		var err error
		if keyword != "" {
			rows, err = db.Query(query, category, "%"+keyword+"%", "%"+keyword+"%")
		} else {
			rows, err = db.Query(query, category)
		}
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		contents := make([]ContentResForHTTPGet, 0)
		for rows.Next() {
			var content ContentResForHTTPGet
			if err := rows.Scan(&content.Id, &content.Title, &content.Description, &content.Url, &content.Image, &content.UploadedBy, &content.Category, &content.Media); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)
				if err := rows.Close(); err != nil {
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			contents = append(contents, content)
		}

		bytes, err := json.Marshal(contents)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
		return

	case http.MethodPost:
		// POSTリクエストの処理
		var req ContentReqForHTTPPost

		// リクエストボディからユーザー情報を読み取る
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			log.Printf("fail: json.Decode, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			log.Println("fail: Title is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//
		if req.Category == "" {
			log.Println("fail: Category is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Media == "" {
			log.Println("fail: Media is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// トランザクションを開始
		tx, err := db.Begin()
		if err != nil {
			log.Printf("fail: db.Begin, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			// リクエスト処理の最後でトランザクションを確定またはロールバックする
			if err != nil {
				log.Printf("fail: Rolling back transaction, %v\n", err)
				if err := tx.Rollback(); err != nil {
					log.Printf("fail: Rollback error, %v\n", err)
				}
			} else {
				if err := tx.Commit(); err != nil {
					log.Printf("fail: Commit error, %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}()

		// ULIDを生成
		Id, err := ulid.New(ulid.Now(), nil)
		if err != nil {
			log.Printf("fail: ulid.New, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		IdStr := Id.String()

		CreateDate := time.Now()

		// データベースに新しいユーザー情報を挿入
		_, err = tx.Exec("INSERT INTO content (id, title, description, url, image, uploaded_by, create_date, category, media) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", IdStr, req.Title, req.Description, req.Url, req.Image, req.UploadedBy, CreateDate, req.Category, req.Media)
		if err != nil {
			log.Printf("fail: tx.Exec, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// トランザクションをコミット
		if err := tx.Commit(); err != nil {
			log.Printf("fail: tx.Commit, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// レスポンスに採番されたULIDを含めて返す
		responseData := map[string]string{"id": IdStr}
		bytes, err := json.Marshal(responseData)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)

	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	http.HandleFunc("/content", handler)
	closeDBWithSysCall()

	log.Println("Listening")
	if err := http.ListenAndServe(":3306", nil); err != nil {
		log.Fatal(err)
	}
}

// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
func closeDBWithSysCall() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)

		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("success: db.Close()")
		os.Exit(0)
	}()
}
