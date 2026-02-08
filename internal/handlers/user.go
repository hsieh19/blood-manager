package handlers

import (
	"fmt"
	"net/http"
	"time"

	"health-manager/internal/database"
	"health-manager/internal/models"

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

// CreateBP 创建健康记录
func CreateBP(c *gin.Context) {
	var req models.CreateBPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据格式错误"})
		return
	}

	// 至少需要填写血压或身高体重之一
	hasBP := req.Systolic > 0 || req.Diastolic > 0
	hasBody := req.Height > 0 || req.Weight > 0

	if !hasBP && !hasBody {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请至少填写血压或身高体重数据"})
		return
	}

	userID := c.GetInt64("user_id")
	recordTime := time.Now().In(beijingLoc)

	id, err := database.CreateBPRecord(userID, req.Systolic, req.Diastolic, req.HeartRate, req.Height, req.Weight, req.Waistline, recordTime, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "记录成功",
		"id":      id,
		"time":    recordTime.Format("2006-01-02 15:04"),
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
