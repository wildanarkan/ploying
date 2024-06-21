package controllers

import (
	"net/http"
	"os"
	"strconv"
	"ploying/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskController struct {
	DB *gorm.DB
}

func (t *TaskController) Create(c *gin.Context) {
	task := models.Task{}
	errBindJson := c.ShouldBindJSON(&task)
	if errBindJson != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errBindJson.Error()})
		return
	}

	errDB := t.DB.Create(&task).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (t *TaskController) Delete(c *gin.Context) {
	id := c.Param("id")
	task := models.Task{}

	if err := t.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	errDB := t.DB.Delete(&models.Task{}, id).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDB.Error()})
		return
	}

	if task.Attachment != "" {
		os.Remove("attachments/" + task.Attachment)
	}

	c.JSON(http.StatusOK, "Deleted")
}

func (t *TaskController) Submit(c *gin.Context) {
	task := models.Task{}
	id := c.Param("id")
	submitDate := c.PostForm("submitDate")
	file, errFile := c.FormFile("attachment")
	if errFile != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errFile.Error()})
		return
	}

	if err := t.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// remove old attachment
	attachment := task.Attachment
	fileInfo, _ := os.Stat("attachments/" + attachment)
	if fileInfo != nil {
		// found
		os.Remove("attachments/" + attachment)
	}

	// create new attaachment
	attachment = file.Filename
	errSave := c.SaveUploadedFile(file, "attachments/"+attachment)
	if errSave != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errSave.Error()})
		return
	}

	errDB := t.DB.Where("id=?", id).Updates(models.Task{
		Status:     "Review",
		SubmitDate: submitDate,
		Attachment: attachment,
	}).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, "Submit to Review")
}

func (t *TaskController) Reject(c *gin.Context) {
	task := models.Task{}
	id := c.Param("id")
	reason := c.PostForm("reason")
	rejectedDate := c.PostForm("rejectedDate")

	if err := t.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	errDB := t.DB.Where("id=?", id).Updates(models.Task{
		Status:       "Rejected",
		Reason:       reason,
		RejectedDate: rejectedDate,
	}).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, "Rejected")
}

func (t *TaskController) Fix(c *gin.Context) {
	id := c.Param("id")
	revision, errConv := strconv.Atoi(c.PostForm("revision"))
	if errConv != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errConv.Error()})
		return
	}

	if err := t.DB.First(&models.Task{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	errDB := t.DB.Where("id=?", id).Updates(models.Task{
		Status:   "Queue",
		Revision: int8(revision),
	}).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, "Fix to Queue")
}

func (t *TaskController) Approve(c *gin.Context) {
	id := c.Param("id")
	approvedDate := c.PostForm("approvedDate")

	if err := t.DB.First(&models.Task{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	errDB := t.DB.Where("id=?", id).Updates(models.Task{
		Status:       "Approved",
		ApprovedDate: approvedDate,
	}).Error
	if errDB != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, "Approved")
}

func (t *TaskController) FindById(c *gin.Context) {
	task := models.Task{}
	id := c.Param("id")

	if err := t.DB.First(&models.Task{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	errDB := t.DB.Preload("User").Find(&task, id).Error
	if errDB != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (t *TaskController) NeedToBeReview(c *gin.Context) {
	tasks := []models.Task{}

	errDB := t.DB.Preload("User").Where("status=?", "Review").Order("submit_date ASC").Limit(2).Find(&tasks).Error
	if errDB != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (t *TaskController) ProgressTasks(c *gin.Context) {
	tasks := []models.Task{}
	userId := c.Param("userId")

	errDB := t.DB.Where(
		"(status!=? AND user_id=?) OR (revision!=? AND user_id=?)", "Queue", userId, 0, userId,
	).Order("updated_at DESC").Limit(5).Find(&tasks).Error
	if errDB != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (t *TaskController) Statistic(c *gin.Context) {
	userId := c.Param("userId")

	stat := []map[string]interface{}{}

	errDB := t.DB.Model(models.Task{}).Select("status, count(status) as total").Where("user_id=?", userId).Group("status").Find(&stat).Error
	if errDB != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, stat)
}

func (t *TaskController) FindByUserAndStatus(c *gin.Context) {
	tasks := []models.Task{}
	userId := c.Param("userId")
	status := c.Param("status")

	errDB := t.DB.Where("user_id=? AND status=?", userId, status).Find(&tasks).Error
	if errDB != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDB.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
