package main

import (
	"context"
	"fmt"
	"net/http"
	"server/db"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type handlerFunc struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (h *handlerFunc) addRecordHandler(c *gin.Context) {
	// 处理添加记录的请求
	param := db.ApplicationRecord{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(400, gin.H{
			"message": "Error: Invalid request body.",
		})
		return
	}
	err := db.AddRecord(param)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error: Failed to add record.",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Record added successfully.",
	})
}

func (h *handlerFunc) removeRecordHandler(c *gin.Context) {
	// 处理删除记录的请求
	appName := c.Param("appName")
	if appName == "" {
		c.JSON(400, gin.H{
			"message": "Error: appName parameter is required.",
		})
		return
	}
	err := db.RemoveRecord(appName)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error: Failed to remove record.",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Record removed successfully.",
	})
}

func (h *handlerFunc) updateOverwriteRecordHandler(c *gin.Context) {
	// 处理更新记录的请求
	param := db.ApplicationRecord{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(400, gin.H{
			"message": "Error: Invalid request body.",
		})
		return
	}
	err := db.UpdateRecord(param)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error: Failed to update record.",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Record updated successfully.",
	})
}

func (h *handlerFunc) updateRecordHandler(c *gin.Context) {
	appName := c.Param("appName")
	part := c.Param("part")
	value := c.Param("value")
	if appName == "" || part == "" || value == "" {
		c.JSON(400, gin.H{
			"message": "Error: appName, part, and value parameters are required.",
		})
		return
	}
	nowRecord, err := db.GetRecord(appName)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error: Failed to get record.",
		})
		return
	}
	var updateErr error
	switch part {
	case "AppLatestVersion":
		ver, updateErr := strconv.Atoi(value)
		nowRecord.AppLatestVersion = ver
		nowRecord.AppAvailableVersion = append(nowRecord.AppAvailableVersion, ver)
		if len(nowRecord.AppAvailableVersion) > 20 {
			nowRecord.AppAvailableVersion = nowRecord.AppAvailableVersion[len(nowRecord.AppAvailableVersion)-20:]
		}
		if updateErr != nil {
			c.JSON(400, gin.H{
				"message": "Error: Invalid value for AppLatestVersion.",
			})
			return
		}
	case "AppForceUpdateMiniumVersion":
		nowRecord.AppForceUpdateMiniumVersion, updateErr = strconv.Atoi(value)
		if updateErr != nil {
			c.JSON(400, gin.H{
				"message": "Error: Invalid value for AppForceUpdateMiniumVersion.",
			})
			return
		}
	case "DirectLink":
		nowRecord.DirectLink = value
		return
	case "NoneDirectLink":
		nowRecord.NoneDirectLink = value
		return
	case "Notice":
		nowRecord.Notice = value
		return
	default:
		c.JSON(400, gin.H{
			"message": "Error: Invalid part parameter.",
		})
		return
	}
}

func (h *handlerFunc) getRecordHandler(c *gin.Context) {
	appName := c.Param("appName")
	if appName == "" {
		c.JSON(400, gin.H{
			"message": "Error: appName parameter is required.",
		})
		return
	}
	record, err := db.GetRecord(appName)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error: Failed to get record.",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Success",
		"data":    record,
	})
}

func (h *handlerFunc) shutdownServerHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Server is shutting down.",
	})
	return
}

func main() {
	fmt.Println("Initializing router...")
	port := fmt.Sprintf(":%d", 8080)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	handlers := &handlerFunc{}
	handlers.ctx, handlers.cancel = context.WithCancel(context.Background())
	// 定义路由
	updaterGroup := r.Group("/api/v1")
	{
		updaterGroup.POST("/add", handlers.addRecordHandler)
		updaterGroup.POST("/remove/:appName", handlers.removeRecordHandler)
		updaterGroup.PUT("/update/overwtite", handlers.updateOverwriteRecordHandler)
		updaterGroup.PUT("/update/part/:appName/:param/:value", handlers.updateRecordHandler)
		updaterGroup.GET("/get/:appName", handlers.getRecordHandler)
	}
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("[ERROR] - Error starting server: %s\n", err.Error())
		}
		return
	}()
	for {
		select {
		case <-handlers.ctx.Done():
			contextShutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			cancelErr := srv.Shutdown(contextShutdown)
			if cancelErr != nil {
				fmt.Println(fmt.Sprintf("[ERROR] - Client HTTP server shutdown error: %s", cancelErr.Error()))
				_ = srv.Close()
				cancel()
				return
			}
			fmt.Println("Client HTTP server shutdown successfully")
			cancel()
			return
		}
	}
}
