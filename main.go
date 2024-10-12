package main

import (
	"log"
	"net"
	"sync"

	"mtg-task/handlers"
	"mtg-task/models"
)

func main() {
	db, err := models.InitDB()
	if err != nil {
		log.Println("[ERR] ошибка инициализации БД:", err)
		return
	}
	defer db.Close()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println("[ERR] ошибка при запуске сервера:", err)
		return
	}
	defer ln.Close()
	log.Println("[INFO] сервер запущен на порту 8080")

	var wg sync.WaitGroup

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("[ERR] ошибка при подключении клиента:", err)
			continue
		}
		wg.Add(1)
		go handlers.HandleConnection(conn, db, &wg)
	}
	wg.Wait()
}
