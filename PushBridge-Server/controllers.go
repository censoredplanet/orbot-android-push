package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/censoredplanet/orbot-android-push/PushBridge-server/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getAllBridges returns all bridges from the database
// @Summary Get all bridges from the database
// @Tags bridges
// @Produce json
// @Router /bridges [get]
// @Success 200 {object} AllBridgeSettingResponse
// @Failure 500 {object} ServerErrorResponse
func getAllBridges(c *gin.Context) {
	// find all countries
	var countries []models.Country

	result := fcmDB.db.Find(&countries)
	if result.Error != nil {
		c.JSON(500, models.ServerErrorResponse{
			Message: "Internal Server Error",
			Error:   result.Error.Error(),
		})
		return
	}

	var allBridges = make(models.AllBridgeSettingResponse, len(countries))
	for _, country := range countries {
		// work around "unexpected end of JSON input" error when field is empty
		if len(country.BridgeSetting) == 0 {
			country.BridgeSetting = "[]"
		}
		allBridges[country.CountryCode] = models.BridgeSettingsResponseFragment{
			Settings: json.RawMessage(country.BridgeSetting),
		}
	}
	c.JSON(200, allBridges)
}

// getBridgesByCountry returns the bridges for a country
// @Summary Get the bridges for a country
// @Tags bridges
// @Produce json
// @Param country path string true "Country Code"
// @Router /bridges/{country} [get]
// @Success 200 {object} BridgeSettingResponse
// @Failure 404 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
func getBridgesByCountry(c *gin.Context) {
	code := c.Param("country")

	var country models.Country
	result := fcmDB.db.First(&country, models.Country{
		CountryCode: code,
	})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(404, models.ServerErrorResponse{
				Message: "Not Found",
				Error:   result.Error.Error(),
			})
			return
		}
		c.JSON(500, models.ServerErrorResponse{
			Message: "Internal Server Error",
			Error:   result.Error.Error(),
		})
		return
	}

	c.JSON(200, models.BridgeSettingResponse{
		Country: code,
		BridgeSettingsResponseFragment: &models.BridgeSettingsResponseFragment{
			Settings: json.RawMessage(country.BridgeSetting),
		},
	})
}

// RegisterFCM registers a user's FCM token and country
// @Summary Register a user's FCM token and country
// @Tags fcm
// @Accept json
// @Produce json
// @Param body body RegisterFCMRequest true "Request Body"
// @Router /fcm/register [post]
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
func registerFCM(c *gin.Context) {
	// cast body to models.RegisterFCMRequest
	var request models.RegisterFCMRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(400, models.ServerErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

	// check if country exists in database, if not, set to "default"
	var country models.Country
	result := fcmDB.db.First(&country, "country_code = ?", strings.ToLower(request.Country))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			country = models.Country{
				CountryCode: "default",
			}
		} else {
			c.JSON(500, models.ServerErrorResponse{
				Message: "Internal Server Error while looking up country",
				Error:   result.Error.Error(),
			})
			return
		}
	}

	// upsert the user
	user := models.User{
		FCMToken: request.FCMToken,
		Country:  country,
	}
	result = fcmDB.db.Create(&user)

	if result.Error != nil {
		c.JSON(500, models.ServerErrorResponse{
			Message: "Internal Server Error while creating user",
			Error:   result.Error.Error(),
		})
		return
	}

	if result.RowsAffected > 0 {
		c.JSON(200, models.MessageResponse{
			Message: "Updated",
		})
	} else {
		c.JSON(200, models.MessageResponse{
			Message: "Already Exists. Not Updated",
		})
	}

	// start a goroutine to send the user the bridges
	go sendBridgeSettingsToUser(user)
}

// TODO: implement this
func updateBridgesUsingMOAT(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// updateBridgesManually updates the bridges for a country manually
// @Summary Update the bridges for a country manually
// @Tags admin
// @Produce json
// @Param country body string true "Country Code"
// @Param settings body string true "Bridge Settings"
// @Router /admin/bridges/set [post]
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
func updateBridgesManually(c *gin.Context) {
	// cast body to models.UpdateBridgesManuallyRequest
	var request models.UpdateBridgesManuallyRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(400, models.ServerErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

	// upsert the country
	country := models.Country{
		CountryCode:   request.Country,
		BridgeSetting: string(request.Settings),
	}
	result := fcmDB.db.FirstOrCreate(&country, models.Country{
		CountryCode: request.Country,
	})
	if result.Error != nil {
		c.JSON(500, models.ServerErrorResponse{
			Message: "Internal Server Error",
			Error:   result.Error.Error(),
		})
		return
	}

	if result.RowsAffected > 0 {
		c.JSON(200, models.MessageResponse{
			Message: "Updated",
		})
	} else {
		c.JSON(200, models.MessageResponse{
			Message: "Already Exists. Not Updated",
		})
	}
}

// notifyFCMByToken sends a notification to a specific user according to their UUID
// @Summary Send a notification to a specific user according to their UUID, containing the bridges for their country
// @Tags admin
// @Produce json
// @Param body body NotifyFCMByTokenRequest true "Request Body"
// @Router /admin/fcm/postUser [post]
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ServerErrorResponse
// @Failure 404 {object} ServerErrorResponse
// @Failure 500 {object} ServerErrorResponse
func notifyFCMByToken(c *gin.Context) {
	// cast body to models.NotifyFCMByTokenRequest
	var request models.NotifyFCMByTokenRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		c.JSON(400, models.ServerErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

	// get the user
	var user models.User
	result := fcmDB.db.First(&user, "id = ?", request.UserID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(404, models.ServerErrorResponse{
				Message: "User Not Found",
				Error:   result.Error.Error(),
			})
			return
		}
		c.JSON(500, models.ServerErrorResponse{
			Message: "Internal Server Error",
			Error:   result.Error.Error(),
		})
		return
	}

	// send the user the bridges
	err = sendBridgeSettingsToUser(user)
	if err != nil {
		c.JSON(500, models.ServerErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(200, models.MessageResponse{
		Message: "Sent",
	})
}

// TODO: implement this. Need to change ORM relationship
func notifyFCMByCountry(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Log the request body
		fmt.Printf("Request Body: %s\n", body)

		// Continue processing the request
		c.Next()
	}
}
