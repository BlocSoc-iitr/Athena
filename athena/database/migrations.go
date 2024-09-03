package database

import (
	"log"
	// "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/DarkLord017/athena/athena/database/models"
)

// func main() {
// 	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
// 	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("Failed to connect to the database: %v", err)
// 	}

// 	// Migrate models
// 	migrateUp(db)
// }

func MigrateUp(db *gorm.DB) {
	// Here you would import and migrate each model like below:
	// This is equivalent to the Python imports in your code

	err := db.AutoMigrate(
		&models.ContractABI{},
		&models.BackfilledRange{},
		&models.Block{},
		&models.DefaultEvent{},
		&models.Transaction{},
	)
	if err != nil {
		log.Fatalf("failed to migrate models: %v", err)
	}

	// If you need to create schemas, GORM doesn't directly support creating schemas like SQLAlchemy,
	// but you can execute raw SQL to do so.

	// Example: Create schema (this is optional and usually not needed in MySQL)
	// db.Exec("CREATE SCHEMA IF NOT EXISTS your_schema_name")
}
