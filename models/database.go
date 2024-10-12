package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Item struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

// Инициализация подключения к БД и создание таблиц
func InitDB() (*sql.DB, error) {

	// Загружаем переменные окружения из файла .env
	err := godotenv.Load("config/config.env")
	if err != nil {
		log.Println("[ERR] ошибка загрузки файла .env")
	}

	// Получаем переменные окружения
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("error connecting to the database:", err)
		return nil, err
	}

	// Создание таблицы items
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100),
        quantity INTEGER,
        price REAL,
        description TEXT
    );`)
	if err != nil {
		log.Println("[ERR] ошибка создания таблицы items:", err)
		return nil, err
	}

	// Создание таблицы data
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS data (
        id TEXT,
        data TEXT
    );`)
	if err != nil {
		log.Println("[ERR] ошибка создания таблицы data:", err)
		return nil, err
	}

	// Заполнение таблицы items данными
	err = SeedData(db)
	if err != nil {
		log.Println("[ERR] error seeding data:", err)
		return nil, err
	}

	return db, nil
}

// Заполнение таблицы items данными
func SeedData(db *sql.DB) error {
	// Проверяем количество записей
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count)
	if err != nil {
		log.Println("[ERR] ошибка при подсчёте записей:", err)
		return err
	}
	if count >= 5000 {
		return nil
	}

	// Очистка таблицы и сброс последовательности
	_, err = db.Exec("TRUNCATE TABLE items RESTART IDENTITY")
	if err != nil {
		log.Println("[ERR] ошибка при очистке таблицы items:", err)
		return err
	}

	// Вставка данных
	tx, err := db.Begin()
	if err != nil {
		log.Println("[ERR] ошибка запуска транзакции:", err)
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO items (name, quantity, price, description) VALUES ($1, $2, $3, $4)")
	if err != nil {
		log.Println("[ERR] ошибка при подготовке запроса:", err)
		return err
	}
	defer stmt.Close()

	for i := 0; i < 5000; i++ {
		name := fmt.Sprintf("Item %d", i+1)
		quantity := i % 100
		price := float64(i) * 0.1
		description := fmt.Sprintf("Description for item %d", i+1)

		_, err = stmt.Exec(name, quantity, price, description)
		if err != nil {
			tx.Rollback()
			log.Println("[ERR] ошибка при вставке данных:", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[ERR] ошибка при коммите транзакции:", err)
		return err
	}

	log.Println("[INFO] таблица items успешно заполнена")
	return nil
}

// Получение данных из таблицы items
func GetItems(db *sql.DB) ([]Item, error) {
	rows, err := db.Query("SELECT id, name, quantity, price, description FROM items")
	if err != nil {
		log.Println("[ERR] ошибка выполнения запроса:", err)
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &item.Price, &item.Description)
		if err != nil {
			log.Println("[ERR] ошибка при чтении строки:", err)
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Println("[ERR] ошибка при обработке строк:", err)
		return nil, err
	}

	return items, nil
}
