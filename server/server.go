package main

import (
	"context"
	"fmt"
	"net/http"
	"server/db"
	"strconv"
	"strings"
	"time"
	log "tools"

	"github.com/gin-gonic/gin"
)

type handlerFunc struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func Contains[T comparable](slice []T, v T) bool {
	for _, x := range slice {
		if x == v {
			return true
		}
	}
	return false
}

func (h *handlerFunc) addRecordHandler(c *gin.Context) {
	// 处理添加记录的请求
	param := db.ApplicationRecordWithoutPath{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(400, gin.H{
			"message": "Error: Invalid request body.",
		})
		return
	}
	full := db.ApplicationRecord{
		AppName:                     param.AppName,
		AppLatestVersion:            param.AppLatestVersion,
		AppForceUpdateMiniumVersion: param.AppForceUpdateMiniumVersion,
		AppAvailableVersion:         param.AppAvailableVersion,
		DirectLink:                  param.DirectLink,
		NoneDirectLink:              param.NoneDirectLink,
		Notice:                      param.Notice,
	}
	err := db.AddRecord(full)
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
	param := db.ApplicationRecordWithoutPath{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(400, gin.H{
			"message": "Error: Invalid request body.",
		})
		return
	}
	full := db.ApplicationRecord{
		AppName:                     param.AppName,
		AppLatestVersion:            param.AppLatestVersion,
		AppForceUpdateMiniumVersion: param.AppForceUpdateMiniumVersion,
		AppAvailableVersion:         param.AppAvailableVersion,
		DirectLink:                  param.DirectLink,
		NoneDirectLink:              param.NoneDirectLink,
		Notice:                      param.Notice,
	}
	err := db.UpdateRecord(full)
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
	part := c.Param("param")
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
	switch strings.ToUpper(part) {
	case "APPLATESTVERSION":
		ver, updateErr := strconv.Atoi(value)
		if updateErr != nil {
			c.JSON(400, gin.H{
				"message": "Error: Invalid value for AppLatestVersion.",
			})
			return
		}
		nowRecord.AppLatestVersion = ver
		if !Contains(nowRecord.AppAvailableVersion, ver) {
			nowRecord.AppAvailableVersion = append(nowRecord.AppAvailableVersion, ver)
		}
		if len(nowRecord.AppAvailableVersion) > 20 {
			nowRecord.AppAvailableVersion = nowRecord.AppAvailableVersion[len(nowRecord.AppAvailableVersion)-20:]
		}
		err := db.UpdateRecord(*nowRecord)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error: Failed to update record.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Record updated successfully.",
		})
		return
	case "APPFORCEUPDATEMINIUMVERSION":
		nowRecord.AppForceUpdateMiniumVersion, updateErr = strconv.Atoi(value)
		if updateErr != nil {
			c.JSON(400, gin.H{
				"message": "Error: Invalid value for AppForceUpdateMiniumVersion.",
			})
			return
		}
		err := db.UpdateRecord(*nowRecord)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error: Failed to update record.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Record updated successfully.",
		})
		return
	case "DIRECTLINK":
		nowRecord.DirectLink = value
		err := db.UpdateRecord(*nowRecord)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error: Failed to update record.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Record updated successfully.",
		})
		return
	case "NONEDIRECTLINK":
		nowRecord.NoneDirectLink = value
		err := db.UpdateRecord(*nowRecord)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error: Failed to update record.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Record updated successfully.",
		})
		return
	case "NOTICE":
		nowRecord.Notice = value
		err := db.UpdateRecord(*nowRecord)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error: Failed to update record.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Record updated successfully.",
		})
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
			"message": fmt.Sprintf("Error: Failed to get record: %s", err.Error()),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Success",
		"data":    record,
	})
}

func (h *handlerFunc) shutdownServerHandler(c *gin.Context) {
	h.cancel()
	c.JSON(200, gin.H{
		"message": "Server is shutting down.",
	})
	return
}

func main() {
	db.InitDB()
	log.Logger.Info("Initializing router...")
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
		updaterGroup.PUT("/update/overwrite", handlers.updateOverwriteRecordHandler)
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
			log.Logger.Error(fmt.Sprintf("Error starting server: %s\n", err.Error()))
		}
		return
	}()
	for {
		select {
		case <-handlers.ctx.Done():
			contextShutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			cancelErr := srv.Shutdown(contextShutdown)
			if cancelErr != nil {
				log.Logger.Error(fmt.Sprintf("Client HTTP server shutdown error: %s", cancelErr.Error()))
				_ = srv.Close()
				cancel()
				return
			}
			log.Logger.Info("Client HTTP server shutdown successfully")
			cancel()
			db.CloseDB()
			return
		}
	}
}
