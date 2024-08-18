package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for simplicity, adjust as needed
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan string)            // Broadcast channel
var mutex = sync.Mutex{}                     // Mutex to protect the clients map
var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

type Product struct {
	ID        int
	Name      string
	Tags      string
	CreatedAt time.Time
}

func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	// Check Redis connection
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Start a Goroutine to subscribe to the "new-products" channel
	go subscribeToChannel(rdb)

	// Start a Goroutine to handle WebSocket broadcast
	go handleMessages()

	// Initialize database connection
	db, err := sql.Open("mysql", "TikTok:your_password@tcp(127.0.0.1:3306)/TikTok_Hackathon")
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer db.Close()

	// Define routes
	http.HandleFunc("/", homePage)
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		submitProduct(w, r, db, rdb)
	})
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		searchProducts(w, r, db)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r)
	})
	http.HandleFunc("/recommendations", func(w http.ResponseWriter, r *http.Request) {
		handleRecommendations(w, r, db)
	})

	// Start server on port 8080
	log.Println("Server starting on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// homePage serves the HTML form
func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// submitProduct handles the form submission, saves the product to MySQL and Redis, and publishes a message to a Redis channel
func submitProduct(w http.ResponseWriter, r *http.Request, db *sql.DB, rdb *redis.Client) {
	if r.Method == http.MethodPost {
		// Parse the form data
		name := r.FormValue("name")
		tags := r.FormValue("tags")

		// Insert the product and tags into the database
		insertQuery := "INSERT INTO TikTok_Shop (name, tags) VALUES (?, ?)"
		_, err := db.Exec(insertQuery, name, tags)
		if err != nil {
			log.Fatal("Failed to insert data into the database:", err)
		}

		// Publish the product details to the Redis channel "new-products"
		err = rdb.Publish(context.Background(), "new-products", fmt.Sprintf("Name: %s, Tags: %s", name, tags)).Err()
		if err != nil {
			log.Fatalf("Failed to publish product to Redis channel: %v", err)
		}

		// Redirect back to the home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// searchProducts handles product search based on tags or product name
func searchProducts(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	queryParam := r.URL.Query().Get("q")
	if queryParam == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Check Redis cache first
	ctx := context.Background()
	cachedData, err := rdb.Get(ctx, queryParam).Result()
	if err == redis.Nil {
		// Cache miss: Query the database
		log.Println("Cache miss, querying database...")

		query := "SELECT id, name, tags, created_at FROM TikTok_Shop WHERE name LIKE ? OR tags LIKE ? ORDER BY created_at DESC"
		rows, err := db.Query(query, "%"+queryParam+"%", "%"+queryParam+"%")
		if err != nil {
			log.Printf("Error executing search query: %v", err)
			http.Error(w, fmt.Sprintf("Failed to execute search query: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var product Product
			var createdAtRaw []uint8

			err := rows.Scan(&product.ID, &product.Name, &product.Tags, &createdAtRaw)
			if err != nil {
				log.Printf("Error scanning product data: %v", err)
				http.Error(w, fmt.Sprintf("Failed to scan product data: %v", err), http.StatusInternalServerError)
				return
			}

			createdAt, err := time.Parse("2006-01-02 15:04:05", string(createdAtRaw))
			if err != nil {
				log.Printf("Error parsing created_at date: %v", err)
				http.Error(w, fmt.Sprintf("Failed to parse created_at date: %v", err), http.StatusInternalServerError)
				return
			}
			product.CreatedAt = createdAt

			products = append(products, product)
		}

		// Cache the result in Redis
		jsonData, err := json.Marshal(products)
		if err == nil {
			rdb.Set(ctx, queryParam, jsonData, 10*time.Minute) // Cache for 10 minutes
		}

		// Return the results to the client
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(products); err != nil {
			log.Printf("Error encoding products as JSON: %v", err)
			http.Error(w, fmt.Sprintf("Failed to encode products as JSON: %v", err), http.StatusInternalServerError)
		}

	} else if err != nil {
		// Error occurred with Redis
		log.Printf("Redis error: %v", err)
		http.Error(w, fmt.Sprintf("Redis error: %v", err), http.StatusInternalServerError)
	} else {
		// Cache hit: Return cached data
		log.Println("Cache hit, returning cached data")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedData))
	}
}

// fetchRecommendedProducts fetches products that match the user's search tags from the TikTok_Shop table
func fetchRecommendedProducts(db *sql.DB, userTags []string) ([]Product, error) {
	query := "SELECT id, name, tags, created_at FROM TikTok_Shop WHERE "

	for i, tag := range userTags {
		if i > 0 {
			query += " OR "
		}
		query += fmt.Sprintf("tags LIKE '%%%s%%'", tag)
	}

	query += " ORDER BY created_at DESC LIMIT 10"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []Product
	for rows.Next() {
		var product Product
		var createdAtRaw []uint8 // Raw data from the DB

		err := rows.Scan(&product.ID, &product.Name, &product.Tags, &createdAtRaw)
		if err != nil {
			return nil, err
		}

		// Convert the raw `created_at` data to a `time.Time`
		createdAt, err := time.Parse("2006-01-02 15:04:05", string(createdAtRaw))
		if err != nil {
			return nil, err
		}
		product.CreatedAt = createdAt

		recommendations = append(recommendations, product)
	}
	return recommendations, nil
}

// handleRecommendations serves recommended products based on user tags as JSON and broadcasts them via WebSocket
func handleRecommendations(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	tags := r.URL.Query().Get("tags")
	if tags == "" {
		http.Error(w, "Query parameter 'tags' is required", http.StatusBadRequest)
		return
	}

	userTags := strings.Split(tags, ",")

	recommendations, err := fetchRecommendedProducts(db, userTags)
	if err != nil {
		http.Error(w, "Failed to fetch recommendations", http.StatusInternalServerError)
		return
	}

	// Send the recommendations back to the client
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recommendations); err != nil {
		http.Error(w, "Failed to encode recommendations as JSON", http.StatusInternalServerError)
	}
}

// broadcastProduct sends a product recommendation to all connected WebSocket clients
func broadcastProduct(product Product) {
	mutex.Lock()
	defer mutex.Unlock()

	productData, err := json.Marshal(product)
	if err != nil {
		log.Printf("Failed to marshal product for broadcast: %v", err)
		return
	}

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, productData)
		if err != nil {
			log.Printf("WebSocket error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// Goroutine to subscribe to the "new-products" channel and process incoming messages
func subscribeToChannel(rdb *redis.Client) {
	// Create a new Redis pub/sub client
	pubsub := rdb.Subscribe(context.Background(), "new-products")

	// Ensure the subscription is ready
	_, err := pubsub.Receive(context.Background())
	if err != nil {
		log.Fatalf("Failed to subscribe to channel: %v", err)
	}

	// Channel to receive messages
	ch := pubsub.Channel()

	log.Println("Subscribed to the new-products channel")

	// Listen for messages
	for msg := range ch {
		log.Printf("Received message from new-products channel: %s\n", msg.Payload)
		broadcast <- msg.Payload
	}
}

// handleWebSocket handles WebSocket requests from clients
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Register the client
	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	log.Println("WebSocket connection established")

	// Listen for WebSocket disconnection
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket disconnected:", err)
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
	}
}

// handleMessages sends messages from the broadcast channel to all connected WebSocket clients
func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		message := <-broadcast

		// Send the message to all connected clients
		mutex.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Printf("WebSocket error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}
