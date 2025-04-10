package controllers

import (
	"context"
	"log"
	"strings"

	//"crypto/rand"
	"fmt"
	//"math"
	"net/http"
	"time"
	"uber-clone/algo"
	"uber-clone/auth"
	"uber-clone/db"
	"uber-clone/models"
	"uber-clone/services"
	"uber-clone/websockets"

	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RequestRide handles the ride request from a rider
func RequestRide(c *gin.Context) {
	var req struct {
		StartLat    float64 `json:"start_lat" binding:"required"`
		StartLng    float64 `json:"start_lng" binding:"required"`
		EndLat      float64 `json:"end_lat" binding:"required"`
		EndLng      float64 `json:"end_lng" binding:"required"`
		VehicleType string  `json:"vehicle_type" binding:"required,oneof=two_wheeler three_wheeler car premium_car"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("✅ Received ride request:", req)

	// Get distance, duration, and fare
	distance, duration, fare, err := services.GetDistance(req.StartLat, req.StartLng, req.EndLat, req.EndLng, req.VehicleType)
	if err != nil {
		fmt.Println("❌ Distance service failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ride details"})
		return
	}
	fmt.Println("✅ Distance calculated:", distance, "km", duration, "mins", "Fare:", fare)

	driverColl := db.GetCollection("drivers")

	var bestDriver models.Driver
	searchRadius := 5000 // Initial radius in meters

	for searchRadius <= 15000 { // Expand search radius up to 15km if no drivers are found
		filter := bson.M{
			"location": bson.M{
				"$nearSphere": bson.M{
					"$geometry": bson.M{
						"type":        "Point",
						"coordinates": []float64{req.StartLng, req.StartLat},
					},
					"$maxDistance": searchRadius,
				},
			},
			"is_available": true,
			"vehicle_type": req.VehicleType,
		}

		var drivers []models.Driver
		fmt.Println("🔍 Searching drivers with radius:", searchRadius)
		cursor, err := driverColl.Find(c, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find drivers"})
			return
		}
		fmt.Println("✅ Mongo query executed, checking cursor...")

		defer cursor.Close(c)

		for cursor.Next(c) {
			var driver models.Driver
			if err := cursor.Decode(&driver); err != nil {
				fmt.Println("⚠️ Decode driver failed:", err)
				continue
			}
			drivers = append(drivers, driver)
		}
		fmt.Printf("🔍 Found %d drivers in %dm radius\n", len(drivers), searchRadius)

		if len(drivers) > 0 {
			bestDriver = drivers[0]
			minDistance := algo.CalculateVincentyDistance(req.StartLat, req.StartLng, bestDriver.Location.Coordinates[1], bestDriver.Location.Coordinates[0])

			for _, driver := range drivers {
				d := algo.CalculateVincentyDistance(req.StartLat, req.StartLng, driver.Location.Coordinates[1], driver.Location.Coordinates[0])
				if d < minDistance {
					minDistance = d
					bestDriver = driver
				}
			}
			break // Stop searching once a driver is found
		}

		searchRadius += 5000 // Increase radius if no drivers are found
	}

	if bestDriver.ID.IsZero() {
		fmt.Println("❌ No drivers found")
		c.JSON(http.StatusNotFound, gin.H{"error": "No available drivers found"})
		return
	}

	// Assign driver and update their status
	update := bson.M{"$set": bson.M{"is_available": false}}
	_, err = driverColl.UpdateOne(c, bson.M{"_id": bestDriver.ID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver availability"})
		return
	}

	otp := generateOTP()
	riderID, _ := primitive.ObjectIDFromHex(c.GetString("user_id"))

	ride := models.Ride{
		RiderID:       riderID,
		StartLocation: models.GeoJSON{Type: "Point", Coordinates: []float64{req.StartLng, req.StartLat}},
		EndLocation:   models.GeoJSON{Type: "Point", Coordinates: []float64{req.EndLng, req.EndLat}},
		Distance:      distance,
		VehicleType:   req.VehicleType,
		Status:        "requested",
		CreatedAt:     time.Now(),
		OTP:           otp,
		DriverID:      bestDriver.ID,
		Fare:          fare,
	}

	rideColl := db.GetCollection("rides")
	result, err := rideColl.InsertOne(c, ride)
	if err != nil {
		fmt.Println("❌ Failed to insert ride:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to request ride"})
		return
	}

	ride.ID = result.InsertedID.(primitive.ObjectID)
	fmt.Println("✅ Ride inserted with ID:", ride.ID.Hex())

	// Notify the driver via WebSocket
	log.Printf("📤 Sending ride request to driver: %s", bestDriver.UserID.Hex())

	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:   "ride_request",
		UserID: bestDriver.UserID.Hex(),

		Payload: gin.H{
			"ride_id":  ride.ID.Hex(),
			"rider_id": ride.RiderID.Hex(),
			"distance": ride.Distance,
			"fare":     ride.Fare,
			"pickup":   ride.StartLocation.Coordinates,
		},
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Ride requested",
		"ride_id":   ride.ID.Hex(),
		"distance":  distance,
		"duration":  duration,
		"fare":      fare,
		"driver_id": bestDriver.ID.Hex(),
		"otp":       otp, // Send OTP for testing
	})
}

func HandleDriverResponse(c *gin.Context) {

	rideIdParam := c.Param("ride_id")

	var req struct {
		// RideID string `json:"ride_id" binding:"required"`
		Accept bool `json:"accept"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rideID, _ := primitive.ObjectIDFromHex(rideIdParam)
	rideColl := db.GetCollection("rides")
	userColl := db.GetCollection("users")
	driverColl := db.GetCollection("drivers")

	update := bson.M{
		"$set": bson.M{
			"status":      "accepted",
			"accepted_at": time.Now(),
		},
	}
	if !req.Accept {
		update = bson.M{
			"$set": bson.M{
				"status":      "rejected",
				"rejected_at": time.Now(),
			},
		}
	}

	_, err := rideColl.UpdateByID(c, rideID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ride status"})
		return
	}

	// Fetch the updated ride data
	var ride models.Ride
	err = rideColl.FindOne(c, bson.M{"_id": rideID}).Decode(&ride)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		return
	}

	// Step 1: Get the driver document
	var driver models.Driver
	err = driverColl.FindOne(c, bson.M{"_id": ride.DriverID}).Decode(&driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Driver not found"})
		return
	}

	// Step 2: Get the user linked to that driver (where the name is)
	var user models.User
	err = userColl.FindOne(c, bson.M{"_id": driver.UserID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Driver user not found"})
		return
	}

	// Now you can use user.Name or whatever
	fmt.Println("Driver Name:", user.Name)

	// Notify the rider via WebSocket
	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:   "ride_response",
		UserID: ride.RiderID.Hex(),
		Payload: gin.H{
			"status":      update["$set"].(bson.M)["status"], // Either "accepted" or "rejected"
			"driver_id":   ride.DriverID.Hex(),
			"driver_name": user.Name,
			"ride_id":     rideID.Hex(),
		},
	}

	c.JSON(http.StatusOK, gin.H{"message": "Response recorded"})
}

// generateOTP generates a six-digit random OTP
func generateOTP() string {
	rand.Seed(time.Now().UnixNano())               // Seed the random number generator
	return fmt.Sprintf("%06d", rand.Intn(1000000)) // Generate a 6-digit OTP
}

// VerifyOTP allows the rider to verify the OTP before starting the ride
func VerifyOTP(c *gin.Context) {
	rideID := c.Param("ride_id")
	var req struct {
		OTP string `json:"otp" binding:"required"`
	}

	// Extract JWT Token from Header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		return
	}

	// Validate the token and extract claims
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Check if the token belongs to a driver
	if claims.Role != "driver" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only drivers can verify OTP"})
		return
	}

	// Bind the OTP from the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	// Find ride from the database using the rideID
	rideColl := db.GetCollection("rides")
	objID, _ := primitive.ObjectIDFromHex(rideID)
	var ride models.Ride
	err = rideColl.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&ride)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		return
	}

	// Fetch the driver details from the drivers table using the driver_id in the ride
	driverColl := db.GetCollection("drivers")
	var driver models.Driver
	err = driverColl.FindOne(context.Background(), bson.M{"_id": ride.DriverID}).Decode(&driver)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Ensure that the user making the request is the driver
	if driver.UserID.Hex() != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to verify OTP for this ride"})
		return
	}

	// Validate OTP
	if ride.OTP != req.OTP {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	// Mark the ride as "ongoing"
	rideColl.UpdateByID(context.Background(), objID, bson.M{"$set": bson.M{"status": "ongoing"}})

	// Send a notification to the rider that the ride has started and is now ongoing
	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:   "ride_started",
		UserID: ride.RiderID.Hex(), // Notify the rider
		Payload: gin.H{
			"ride_id": rideID,
			"message": "Your ride has started and is now ongoing.",
		},
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified, ride is now ongoing"})
}

func CancelRide(c *gin.Context) {
	rideID := c.Param("ride_id")
	var req struct {
		Reason string `json:"reason"`
	}

	// Extract JWT Token from Header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		return
	}

	// Validate the token and extract claims
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Convert rideID to ObjectID
	objID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ride ID"})
		return
	}

	// Find ride from the database using the rideID
	rideColl := db.GetCollection("rides")
	var ride models.Ride
	err = rideColl.FindOne(c, bson.M{"_id": objID}).Decode(&ride)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		return
	}

	// Check if the ride is already completed and prevent cancellation
	if ride.Status == "completed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Ride has already been completed and cannot be cancelled"})
		return
	}

	// Check if the user is either the rider or the driver
	userID := claims.UserID
	if ride.RiderID.Hex() != userID {
		// Check if the user is the driver, but we need to compare with the Driver's collection
		driverColl := db.GetCollection("drivers")
		var driver models.Driver
		err = driverColl.FindOne(c, bson.M{"user_id": userID}).Decode(&driver)
		if err != nil || ride.DriverID.Hex() != driver.ID.Hex() {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to cancel this ride"})
			return
		}
	}

	// Update the ride status to "cancelled"
	update := bson.M{
		"$set": bson.M{
			"status":       "cancelled",
			"cancelled_by": userID, // Store the user who cancelled the ride
			"cancelled_at": time.Now(),
			"reason":       req.Reason, // Add reason if provided
		},
	}

	// Perform the update in the database
	_, err = rideColl.UpdateByID(c, objID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel the ride"})
		return
	}

	// Send a notification to the rider
	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:    "ride_cancelled",
		UserID:  ride.RiderID.Hex(), // Notify the rider
		Payload: gin.H{"ride_id": rideID, "message": "Your ride has been cancelled."},
	}

	// Get the driver’s user_id by looking up the driver in the drivers collection
	driverColl := db.GetCollection("drivers")
	var driver models.Driver
	err = driverColl.FindOne(c, bson.M{"_id": ride.DriverID}).Decode(&driver)
	if err != nil {
		fmt.Println("Driver not found or error fetching driver details:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding driver for notification"})
		return
	}

	// Now that we have the driver's user_id, send the notification to the driver
	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:    "ride_cancelled",
		UserID:  driver.UserID.Hex(), // Use the user_id from the drivers collection
		Payload: gin.H{"ride_id": rideID, "message": "The ride has been cancelled."},
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ride cancelled"})
}

func SubmitFeedback(c *gin.Context) {
	var req struct {
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
	}

	// Bind the request body to the 'req' struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract RideID and UserID from the context or request parameters
	rideID := c.Param("ride_id")
	userID := c.GetString("user_id")

	// Convert the extracted RideID to ObjectID
	objID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ride ID"})
		return
	}

	// Convert the extracted UserID to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	// Create the feedback document
	feedback := models.Feedback{
		RideID:    objID,
		UserID:    userObjID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: time.Now(),
	}

	// Insert feedback into the "feedback" collection in MongoDB
	feedbackColl := db.GetCollection("feedback")
	_, err = feedbackColl.InsertOne(c, feedback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	existingFeedback := feedbackColl.FindOne(c, bson.M{
		"ride_id": objID,
		"user_id": userObjID,
	})
	if existingFeedback.Err() == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already submitted feedback for this ride"})
		return
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{"message": "Feedback submitted"})
}

func CompleteRide(c *gin.Context) {
	rideID := c.Param("ride_id") // Ride ID passed as a URL parameter

	// Extract JWT token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		return
	}

	// Validate the token and extract claims
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Convert rideID to ObjectID
	rideObjID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ride ID"})
		return
	}

	// Get ride details from the database
	rideColl := db.GetCollection("rides")
	var ride models.Ride
	err = rideColl.FindOne(c, bson.M{"_id": rideObjID}).Decode(&ride)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		return
	}

	// Fetch the driver details from the drivers table using the driver_id in the ride
	driverColl := db.GetCollection("drivers")
	var driver models.Driver
	err = driverColl.FindOne(c, bson.M{"_id": ride.DriverID}).Decode(&driver)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Ensure that the user making the request is the driver
	if driver.UserID.Hex() != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to complete this ride"})
		return
	}

	// Update the ride status to "completed"
	update := bson.M{
		"$set": bson.M{
			"status":       "completed",
			"completed_at": time.Now(),
		},
	}

	_, err = rideColl.UpdateByID(c, rideObjID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete the ride"})
		return
	}

	// Send a notification to the rider and driver about the ride completion
	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:    "ride_completed",
		UserID:  ride.RiderID.Hex(), // Notify the rider
		Payload: gin.H{"ride_id": rideObjID.Hex()},
	}

	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:    "ride_completed",
		UserID:  driver.UserID.Hex(), // Notify the driver
		Payload: gin.H{"ride_id": rideObjID.Hex()},
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ride completed"})
}

func HandlePayment(c *gin.Context) {
	rideID := c.Param("ride_id")
	userID := c.GetString("user_id")

	// Convert IDs to ObjectID
	rideObjID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ride ID"})
		return
	}

	// Get ride details
	rideColl := db.GetCollection("rides")
	var ride models.Ride
	if err := rideColl.FindOne(c, bson.M{"_id": rideObjID}).Decode(&ride); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		return
	}

	// Verify user is the rider
	if ride.RiderID.Hex() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only rider can pay for this ride"})
		return
	}

	// Check if the ride has already been paid for (check payment_status)
	if ride.PaymentStatus != "" { // If the payment_status is not empty, payment has already been made
		c.JSON(http.StatusConflict, gin.H{"error": "Payment already made for this ride"})
		return
	}

	// Check if payment already exists in the payments collection
	paymentColl := db.GetCollection("payments")
	var existingPayment models.Payment
	err = paymentColl.FindOne(c, bson.M{"ride_id": rideObjID}).Decode(&existingPayment)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Payment already exists for this ride"})
		return
	}

	if ride.Fare < 50 {
		ride.Fare = 50 // Set the fare to ₹50 if it's less than ₹50
	}

	// Create Stripe Payment Intent
	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(int64(ride.Fare * 100)), // Convert to cents
		Currency:           stripe.String(string(stripe.CurrencyINR)),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	}

	// Add metadata using the AddMetadata method
	params.AddMetadata("ride_id", rideID)
	params.AddMetadata("user_id", userID)

	pi, err := paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment intent"})
		return
	}

	// Save payment record
	payment := models.Payment{
		RideID:        rideObjID,
		Amount:        ride.Fare,
		Currency:      "INR",
		PaymentIntent: pi.ID,
		Status:        "requires_payment_method",
		CreatedAt:     time.Now(),
	}

	if _, err := paymentColl.InsertOne(c, payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save payment record"})
		return
	}

	// Update the ride payment status to "pending" or similar, indicating payment is requested
	updateRideResult, err := rideColl.UpdateOne(
		c,
		bson.M{"_id": rideObjID},
		bson.M{
			"$set": bson.M{
				"payment_status": "pending",
			},
		},
	)
	if err != nil || updateRideResult.MatchedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ride payment status"})
		return
	}

	// Get the driver's details from the drivers collection
	driverColl := db.GetCollection("drivers")
	var driver models.Driver
	if err := driverColl.FindOne(c, bson.M{"_id": ride.DriverID}).Decode(&driver); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Now, retrieve the actual user's details from the users collection using the driver.user_id
	userColl := db.GetCollection("users")
	var driverUser models.User
	if err := userColl.FindOne(c, bson.M{"_id": driver.UserID}).Decode(&driverUser); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver user not found"})
		return
	}

	// Send WebSocket notifications to the rider and driver about the payment request
	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:   "payment_requested",
		UserID: ride.RiderID.Hex(), // Notify the rider
		Payload: gin.H{
			"ride_id":  rideObjID.Hex(),
			"amount":   ride.Fare,
			"currency": "INR",
		},
	}

	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:   "payment_requested",
		UserID: driverUser.ID.Hex(), // Notify the driver
		Payload: gin.H{
			"ride_id":  rideObjID.Hex(),
			"amount":   ride.Fare,
			"currency": "INR",
		},
	}

	// Respond with the payment details
	c.JSON(http.StatusCreated, gin.H{
		"client_secret": pi.ClientSecret,
		"payment_id":    pi.ID,
		"amount":        ride.Fare,
		"currency":      "INR",
	})
}

func ConfirmPayment(c *gin.Context) {
	rideID := c.Param("ride_id")
	var req struct {
		PaymentIntentID string `json:"payment_intent_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get payment intent details from Stripe
	pi, err := paymentintent.Get(req.PaymentIntentID, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment intent"})
		return
	}

	// Verify payment succeeded
	if pi.Status != stripe.PaymentIntentStatusSucceeded {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment not completed"})
		return
	}

	// Update payment record
	paymentColl := db.GetCollection("payments")
	update := bson.M{
		"$set": bson.M{
			"status":       "succeeded",
			"completed_at": time.Now(),
		},
	}

	_, err = paymentColl.UpdateOne(c,
		bson.M{"payment_intent": req.PaymentIntentID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// Update ride status
	rideColl := db.GetCollection("rides")
	rideObjID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ride ID"})
		return
	}

	// Retrieve the ride from the rides collection
	var ride models.Ride
	if err := rideColl.FindOne(c, bson.M{"_id": rideObjID}).Decode(&ride); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		return
	}

	// Now, we update the ride status and payment status
	_, err = rideColl.UpdateOne(c,
		bson.M{"_id": rideObjID},
		bson.M{"$set": bson.M{
			"payment_status": "paid",
			"status":         "completed",
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ride status"})
		return
	}

	// Retrieve the driver's user details from the drivers collection
	driverColl := db.GetCollection("drivers")
	var driver models.Driver
	if err := driverColl.FindOne(c, bson.M{"_id": ride.DriverID}).Decode(&driver); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Get the user details of the driver from the users collection
	userColl := db.GetCollection("users")
	var driverUser models.User
	if err := userColl.FindOne(c, bson.M{"_id": driver.UserID}).Decode(&driverUser); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver user not found"})
		return
	}

	// Now, update the driver's availability status
	_, err = driverColl.UpdateOne(c,
		bson.M{"_id": driver.ID},
		bson.M{"$set": bson.M{
			"is_available": true, // Set the driver as available
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver's availability"})
		return
	}

	// Send WebSocket notifications to the rider and driver about the payment confirmation
	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:   "payment_confirmed",
		UserID: ride.RiderID.Hex(),
		Payload: gin.H{
			"ride_id":  rideObjID.Hex(),
			"amount":   ride.Fare,
			"currency": "INR",
		},
	}

	websockets.WS_HUB.Broadcast <- websockets.Notification{
		Type:   "payment_confirmed",
		UserID: driverUser.ID.Hex(),
		Payload: gin.H{
			"ride_id":  rideObjID.Hex(),
			"amount":   ride.Fare,
			"currency": "INR",
		},
	}

	// Respond with the success message
	c.JSON(http.StatusOK, gin.H{"message": "Payment confirmed successfully"})
}

func GetRideDetails(c *gin.Context) {
	rideID := c.Param("ride_id")

	// Convert rideID and userID to ObjectID
	rideObjID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ride ID"})
		return
	}
	rideColl := db.GetCollection("rides")
	var ride models.Ride
	err = rideColl.FindOne(c, bson.M{"_id": rideObjID}).Decode(&ride)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             ride.ID,
		"rider_id":       ride.RiderID,
		"driver_id":      ride.DriverID,
		"pickup":         ride.StartLocation,
		"destination":    ride.EndLocation,
		"status":         ride.Status,
		"fare":           ride.Fare,
		"payment_status": ride.PaymentStatus, // Paid, Pending
		"created_at":     ride.CreatedAt,
	})
}
