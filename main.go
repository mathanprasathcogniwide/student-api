package main

import (
	"database/sql"

	"log"

	"net"

	"net/http"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"

	_ "github.com/lib/pq"
)

var db *gorm.DB

type Student struct {
	Name string `json:"name" gorm:"column:name"`

	RollNo string `json:"roll_no" gorm:"primaryKey;column:roll_no"`

	Department string `json:"department" gorm:"column:department"`
}

func createDBIfNotExists() {

	psqlInfo := "user=postgres host=project-db.c5qmq2w6e14z.us-east-1.rds.amazonaws.com password=123456789  port=5432 dbname=studentdb sslmode=require"

	sqlDB, err := sql.Open("postgres", psqlInfo)

	if err != nil {

		log.Fatal("Failed to connect to PostgreSQL:", err)

	}

	defer sqlDB.Close()

	var exists bool

	err = sqlDB.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'studentdb')").Scan(&exists)

	if err != nil {

		log.Fatal("Failed to check if database exists:", err)

	}

	if !exists {

		_, err = sqlDB.Exec("CREATE DATABASE studentdb")

		if err != nil {

			log.Fatal("Failed to create database:", err)

		}

		log.Println("Database 'studentdb' created successfully")

	} else {

		log.Println("Database 'studentdb' already exists")

	}

}

func initDB() {

	createDBIfNotExists()

	dsn := "user=postgres host=project-db.c5qmq2w6e14z.us-east-1.rds.amazonaws.com password=123456789  port=5432 dbname=studentdb sslmode=require"

	var err error

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {

		log.Fatal("Failed to connect to studentdb:", err)

	}

	if err := db.AutoMigrate(&Student{}); err != nil {

		log.Fatal("Failed to migrate database:", err)

	}

}

func getStudents(c *gin.Context) {

	var students []Student

	if result := db.Find(&students); result.Error != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving students"})

		return

	}

	c.JSON(http.StatusOK, students)

}

func getStudent(c *gin.Context) {

	rollNo := c.Param("roll_no")

	var student Student

	if result := db.First(&student, "roll_no = ?", rollNo); result.Error != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})

		return

	}

	c.JSON(http.StatusOK, student)

}

func createStudent(c *gin.Context) {

	var student Student

	if err := c.ShouldBindJSON(&student); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})

		return

	}

	if result := db.Create(&student); result.Error != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating student"})

		return

	}

	c.JSON(http.StatusCreated, gin.H{"message": "Student created successfully", "student": student})

}

func updateStudent(c *gin.Context) {

	rollNo := c.Param("roll_no")

	var student Student

	if result := db.First(&student, "roll_no = ?", rollNo); result.Error != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})

		return

	}

	if err := c.ShouldBindJSON(&student); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})

		return

	}

	if result := db.Save(&student); result.Error != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating student"})

		return

	}

	c.JSON(http.StatusOK, gin.H{"message": "Student updated successfully", "student": student})

}

func deleteStudent(c *gin.Context) {

	rollNo := c.Param("roll_no")

	if result := db.Delete(&Student{}, "roll_no = ?", rollNo); result.Error != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting student"})

		return

	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Student deleted successfully"})

}

func main() {

	initDB()

	router := gin.Default()

	router.GET("/students", getStudents)

	router.GET("/students/:roll_no", getStudent)

	router.POST("/students", createStudent)

	router.PUT("/students/:roll_no", updateStudent)

	router.DELETE("/students/:roll_no", deleteStudent)

	router.GET("/health", func(c *gin.Context) {

		c.String(http.StatusOK, "Healthy")

	})

	log.Println("Starting server on port 8084...")

	listener, err := net.Listen("tcp4", ":8084")

	if err != nil {

		log.Fatalf("Error starting server: %v", err)

	}

	if err := router.RunListener(listener); err != nil {

		log.Fatalf("Error starting server: %v", err)
	}
}
