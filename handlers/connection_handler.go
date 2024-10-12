package handlers

import (
	"bufio"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"mtg-task/models"
)

type Item struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

// Обработка соединения с клиентом
func HandleConnection(conn net.Conn, db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	clientID := conn.RemoteAddr().String()
	log.Printf("[INFO] клиент подключился: %v (ID: %s)\n", conn.RemoteAddr(), clientID)

	reader := bufio.NewReader(conn)
	for {
		// Читаем запрос от клиента
		request, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("[INFO] клиент отключился: %v (ID: %s)\n", conn.RemoteAddr(), clientID)
			} else {
				log.Println("[ERR] ошибка чтения от клиента:", err)
			}
			break
		}

		request = strings.TrimSpace(request)

		if request == "GET_DATA" {
			// Обработка запроса на получение данных
			err := sendDataToClient(conn, db)
			if err != nil {
				log.Println("[ERR] ошибка отправки данных клиенту:", err)
			}
		} else {
			handleClientData(request, clientID, db)
		}
	}
}

// Отправка данных клиенту
func sendDataToClient(conn net.Conn, db *sql.DB) error {
	// Получение данных из базы данных
	items, err := models.GetItems(db)
	if err != nil {
		log.Println("[ERR] ошибка получения данных из БД:", err)
		return err
	}

	writer := bufio.NewWriter(conn)

	for _, item := range items {
		// Кодирование данных в JSON
		jsonData, err := json.Marshal(item)
		if err != nil {
			log.Println("[ERR] ошибка кодирования JSON:", err)
			continue
		}

		// Кодирование в Base64
		encodedData := base64.StdEncoding.EncodeToString(jsonData)

		// Отправка данных клиенту
		_, err = writer.WriteString(encodedData + "\n")
		if err != nil {
			log.Println("[ERR] ошибка отправки данных клиенту:", err)
			continue
		}

		// Сброс буфера для отправки данных
		err = writer.Flush()
		if err != nil {
			log.Println("[ERR] ошибка при отправке данных клиенту:", err)
			continue
		}
	}

	return nil
}

// Получаем данные от клиента для записи в БД
func handleClientData(encodedData string, clientID string, db *sql.DB) {
	// Декодируем данные из Base64
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		log.Println("[ERR] ошибка декодирования Base64:", err)
		return
	}

	// Сохранение данных в базу
	_, err = db.Exec("INSERT INTO data (id, data) VALUES ($1, $2)", clientID, string(decodedData))
	if err != nil {
		log.Println("[ERR] ошибка записи в базу данных:", err)
		return
	}

	log.Printf("[INFO] получены новые данные от ID: %s\n", clientID)
}
