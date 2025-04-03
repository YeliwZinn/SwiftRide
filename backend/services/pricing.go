package services

import (
    "context"
    "math"
    "uber-clone/db"
    "go.mongodb.org/mongo-driver/bson"
)

// CalculateSurge returns a surge multiplier based on demand/supply
func CalculateSurge() (float64, error) {
    rideColl := db.GetCollection("rides")
    driverColl := db.GetCollection("drivers")

    // Demand: Number of active ride requests
    demand, _ := rideColl.CountDocuments(context.Background(), bson.M{
        "status": bson.M{"$in": []string{"requested", "ongoing"}},
    })

    // Supply: Number of available drivers
    supply, _ := driverColl.CountDocuments(context.Background(), bson.M{"is_available": true})

    if supply == 0 {
        return 3.0, nil // Max surge if no drivers
    }

    surge := 1.0 + (float64(demand)/float64(supply)) * 0.5
    return math.Min(surge, 3.0), nil // Cap surge at 3x
}

// CalculateFare computes the total fare
func CalculateFare(distance float64, vehicleType string, surge float64) float64 {
    baseRate := map[string]float64{
        "two_wheeler":   0.8,
        "three_wheeler": 1.2,
        "car":           1.5,
        "premium_car":   2.5,
    }

    // Ensure a valid vehicle type
    if rate, exists := baseRate[vehicleType]; exists {
        return rate * distance * surge
    }
    return 0.0 // Invalid vehicle type
}