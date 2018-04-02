package main

import "net/http"
import "database/sql"

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type DbExplorer struct {
	db *sql.DB
}

func (db DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("everything is OK!"))
}



func NewDbExplorer(db *sql.DB) (DbExplorer, error) {
	return DbExplorer{db}, nil
}