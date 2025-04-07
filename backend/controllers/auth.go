package controllers

import (
	"context"
	"log"
	"net/http"
	"time"
	"uber-clone/auth"
	"uber-clone/db"
	"uber-clone/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Signup handles user registration
func Signup(c *gin.Context) {
	var req struct {
		Name          string  `json:"name" binding:"required"`
		Email         string  `json:"email" binding:"required,email"`
		Phone         string  `json:"phone" binding:"required"`
		Password      string  `json:"password" binding:"required,min=6"`
		Role          string  `json:"role" binding:"required,oneof=rider driver"`
		VehicleType   string  `json:"vehicle_type,omitempty"` // Only for drivers
		LicenseNumber string  `json:"license_number,omitempty"`
		CarPlate      string  `json:"car_plate,omitempty"`
		Lat           float64 `json:"lat" binding:"required"`
		Lng           float64 `json:"lng" binding:"required"`
	}

	// Bind the incoming JSON request to the struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate if role-specific fields are provided when the role is 'driver'
	if req.Role == "driver" {
		if req.VehicleType == "" || req.LicenseNumber == "" || req.CarPlate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All driver-related fields must be provided"})
			return
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create the user object
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Role:     req.Role,
		Location: models.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{req.Lng, req.Lat},
		},
		CreatedAt: time.Now().UTC(),
	}

	// Insert user into the database
	userColl := db.GetCollection("users")
	result, err := userColl.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// If the user is a driver, create a driver profile
	if req.Role == "driver" {
		insertedID, ok := result.InsertedID.(primitive.ObjectID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create driver profile"})
			return
		}

		// Create the driver profile with the same location as the user
		driver := models.Driver{
			UserID:        insertedID,
			VehicleType:   req.VehicleType,
			LicenseNumber: req.LicenseNumber,
			CarPlate:      req.CarPlate,
			IsAvailable:   true, // All drivers start as available
			Location: models.GeoJSON{
				Type:        "Point",
				Coordinates: []float64{req.Lng, req.Lat},
			},
			CreatedAt: time.Now().UTC(),
		}

		driverColl := db.GetCollection("drivers")
		_, err := driverColl.InsertOne(context.Background(), driver)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create driver profile"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// Login handles user authentication
func Login(c *gin.Context) {
	var req struct {
		Email    string  `json:"email" binding:"required,email"`
		Password string  `json:"password" binding:"required,min=6"`
		Lat      float64 `json:"lat" binding:"required"`
		Lng      float64 `json:"lng" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	userColl := db.GetCollection("users")
	var user models.User
	err := userColl.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update location and last_login
	update := bson.M{
		"$set": bson.M{
			"location": models.GeoJSON{
				Type:        "Point",
				Coordinates: []float64{req.Lng, req.Lat},
			},
			"last_login": time.Now(),
		},
	}
	userID, err := primitive.ObjectIDFromHex(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	result, err := userColl.UpdateByID(c, userID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location and last login"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found to update"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID.Hex(), user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func Profile(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		log.Println("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDValue.(string))
	if err != nil {
		log.Println("Invalid user ID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	userColl := db.GetCollection("users")
	var user models.User
	err = userColl.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		log.Println("User not found in database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	log.Println("User details fetched successfully:", user.Email)

	// Extract coordinates safely
	var longitude, latitude float64
	if len(user.Location.Coordinates) == 2 {
		longitude = user.Location.Coordinates[0]
		latitude = user.Location.Coordinates[1]
	}

	c.JSON(http.StatusOK, gin.H{
		"name":      user.Name,
		"email":     user.Email,
		"phone":     user.Phone,
		"role":      user.Role,
		"longitude": longitude,
		"latitude":  latitude,
	})
}
