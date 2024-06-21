package config

import (
	"fmt"
	"ploying/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	host     = "localhost"
	port     = 3306
	user     = "root"
	password = ""
	dbName   = "ploying"
)

func DatabaseConnection() *gorm.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		user, password, host, port, dbName,
	)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return database
}

func CreateOwnerAccount(db *gorm.DB) {
	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	owner := models.User{
		Role:     "Admin",
		Name:     "Owner",
		Password: string(hashedPasswordBytes),
		Email:    "owner@go.id",
	}

	if db.Where("email=?", owner.Email).First(&owner).RowsAffected == 0 {
		db.Create(&owner)
	} else {
		fmt.Println("owner exist")
	}

}
