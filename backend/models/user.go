package models

import (
	"time"
	//"uber-clone/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// models/user.go
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email" unique:"true"`
	Phone     string             `bson:"phone"`
	Password  string             `bson:"password"`
	Role      string             `bson:"role"` // /rider/driver
	Location  GeoJSON            `bson:"location"`
	CreatedAt time.Time          `bson:"created_at"`
}

// models/driver.go
type Driver struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	UserID        primitive.ObjectID `bson:"user_id"`
	VehicleType   string             `bson:"vehicle_type"`
	LicenseNumber string             `bson:"license_number"`
	CarPlate      string             `bson:"car_plate"`
	IsAvailable   bool               `bson:"is_available"`
	Location      GeoJSON            `bson:"location"`
	CreatedAt     time.Time          `json:"created_at"`
}

type Ride struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	RiderID         primitive.ObjectID `bson:"rider_id"`  // Reference to Users
	DriverID        primitive.ObjectID `bson:"driver_id"` // Reference to Drivers
	StartLocation   GeoJSON            `bson:"start_loc"` // GeoJSON Point
	EndLocation     GeoJSON            `bson:"end_loc"`   // GeoJSON Point
	Distance        float64            `bson:"distance"`  // In km (from DistanceMatrix.ai Maps)
	Fare            float64            `bson:"fare"`      // Final calculated fare
	VehicleType     string             `bson:"vehicle_type" validate:"required,oneof=two_wheeler three_wheeler car premium_car"`
	Status          string             `bson:"status" validate:"oneof=requested ongoing completed cancelled"`
	OTP             string             `bson:"otp"` // 6-digit code
	SurgeMultiplier float64            `bson:"surge" default:"1.0"`
	CancelledBy     string             `bson:"cancelled_by" validate:"omitempty,oneof=rider driver"`
	CancellationFee float64            `bson:"cancellation_fee" default:"0"`
	CreatedAt       time.Time          `bson:"created_at"`
	CancelledAt     time.Time          `bson:"cancelled_at,omitempty"`
	AcceptedAt      time.Time          `bson:"accepted_at,omitempty"`
	RejectedAt      time.Time          `bson:"rejected_at,omitempty"`
	CompletedAt     time.Time          `bson:"completed_at,omitempty"`
	PaymentStatus   string             `bson:"payment_status" default:"pending"`
}

type GeoJSON struct {
	Type        string    `bson:"type" default:"Point"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [longitude, latitude]
}

type Payment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	RideID        primitive.ObjectID `bson:"ride_id"`        // Reference to Rides
	Amount        float64            `bson:"amount"`         // Positive: Rider paid, Negative: Refund
	Currency      string             `bson:"currency"`       // Currency code, e.g., "INR"
	PaymentIntent string             `bson:"payment_intent"` // Stripe Payment Intent ID
	Status        string             `bson:"status" validate:"oneof=pending completed failed"`
	CreatedAt     time.Time          `bson:"created_at"`
	StripeID      string             `bson:"stripe_id"`
}

type Feedback struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	RideID    primitive.ObjectID `bson:"ride_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Rating    int                `bson:"rating" validate:"min=1,max=5"`
	Comment   string             `bson:"comment"`
	CreatedAt time.Time          `bson:"created_at"`
}
