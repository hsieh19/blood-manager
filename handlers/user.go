package handlers

import (
	"fmt"
	"net/http"
	"time"

	"blood-manager/database"
	"blood-manager/models"

	"github.com/gin-gonic/gin"
)

// 北京时区
var beijingLoc *time.Location

func init() {
	var err error
	beijingLoc, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		beijingLoc = time.FixedZone("CST", 8*3600)
	}
}

// CreateBP 创建血压记录
func CreateBP(c *gin.Context) {
	var req models.CreateBPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整的血压数据"})
		return
	}

	userID := c.GetInt64("user_id")
	recordTime := time.Now().In(beijingLoc)

	id, err := database.CreateBPRecord(userID, req.Systolic, req.Diastolic, req.HeartRate, recordTime, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "记录成功",
		"id":      id,
		"time":    recordTime.Format("2006-01-02 15:04:05"),
	})
}

// GetBPRecords 获取血压记录
func GetBPRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	records, err := database.GetBPRecords(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 转换时间为北京时间显示
	for i := range records {
		records[i].RecordTime = records[i].RecordTime.In(beijingLoc)
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}

// DeleteBP 删除血压记录
func DeleteBP(c *gin.Context) {
	userID := c.GetInt64("user_id")
	bpID := c.Param("id")

	var id int64
	fmt.Sscanf(bpID, "%d", &id)

	if err := database.DeleteBPRecord(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
