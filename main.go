// desenvolvido por GABRIEL FERRANDIN
package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "encurtador"
)

var db *sql.DB

func main() {
	// Configurar a conexão com o banco de dados PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Testar a conexão com o banco de dados
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Inicializar o roteador Gin
	router := gin.Default()

	// Definir uma rota POST para encurtar URLs
	router.POST("/", ShortenURL)

	//Definir rota get para retornar url pelo codigo
	router.GET("/", GetCodeToUrl)

	// Executar o servidor na porta 8080
	router.Run(":8080")
}

// Função para gerar um código alfanumérico com 6 dígitos
func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// ShortenURL é a função que encurtará a URL
func ShortenURL(c *gin.Context) {
	// Obter a URL recebida no corpo da requisição
	var request struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Gerar um código único com 6 dígitos
	shortCode := generateShortCode()
	fmt.Println(shortCode)

	// Inserir a URL original e o código no banco de dados
	_, err := db.Exec("INSERT INTO urls (original_url, short_code) VALUES ($1, $2)", request.URL, shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao encurtar a URL"})
		return
	}

	// Retornar o código único como resposta
	c.JSON(http.StatusOK, gin.H{"short_code": shortCode})
}

func GetCodeToUrl(c *gin.Context) {
	var request struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Realize uma consulta no banco de dados para obter a URL correspondente ao código
	var originalURL string
	err := db.QueryRow("SELECT original_url FROM urls WHERE short_code = $1", request.Code).Scan(&originalURL)

	if err != nil {
		// Se o código não for encontrado no banco de dados, retorne um erro
		c.JSON(http.StatusNotFound, gin.H{"error": "Código não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"original_url": originalURL})
}
