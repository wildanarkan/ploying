package controllers

import (
	"net/http"
	"ploying/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func (u *UserController) Login(c *gin.Context) {
	user := models.User{}
	errBindJson := c.ShouldBindJSON(&user)
	if errBindJson != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errBindJson.Error()})
		return
	}

	password := user.Password

	errDB := u.DB.Where("email=?", user.Email).Take(&user).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or Password is Wrong"})
		return
	}

	errHash := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if errHash != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email or Password is Wrong"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (u *UserController) CreateAccount(c *gin.Context) {
	user := models.User{}
	errBindJson := c.ShouldBindJSON(&user)
	if errBindJson != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errBindJson.Error()})
		return
	}

	emailExist := u.DB.Where("email=?", user.Email).First(&user).RowsAffected != 0
	if emailExist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exist"})
		return
	}

	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	user.Password = string(hashedPasswordBytes)
	user.Role = "Employee"

	errDB := u.DB.Create(&user).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (u *UserController) Delete(c *gin.Context) {
	id := c.Param("id")

	errDB := u.DB.Delete(&models.User{}, id).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, "Deleted")
}

func (u *UserController) GetEmployee(c *gin.Context) {
	users := []models.User{}

	errDB := u.DB.Select("id,name,email").Where("role=?", "Employee").Find(&users).Error
	if errDB != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (u *UserController) UpdatePassword(c *gin.Context) {
    type UpdatePasswordInput struct {
        NewPassword string `json:"new_password" binding:"required"`
    }

    var input UpdatePasswordInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userId := c.Param("id")
    var user models.User
    if err := u.DB.Where("id = ?", userId).First(&user).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
        return
    }

    user.Password = string(hashedPassword)
    if err := u.DB.Save(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}