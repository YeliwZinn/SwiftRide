
package services

import (
    "encoding/json"
    "fmt"
    "net/http"
    "uber-clone/config"
)

// Mapbox Directions API Response Structure
type DirectionsResponse struct {
    Routes []struct {
        Distance float64 `json:"distance"` // in meters
        Duration float64 `json:"duration"` // in seconds
        Geometry struct {
            Coordinates [][]float64 `json:"coordinates"`
        } `json:"geometry"`
    } `json:"routes"`
    Code    string `json:"code"`
}

// GetDistance calculates distance and duration using Mapbox Directions API
func GetDistance(originLat, originLng, destLat, destLng float64, vehicleType string) (float64, float64, float64, error) {
    apiKey := config.MustGetEnv("MAPBOX_ACCESS_TOKEN")
    url := fmt.Sprintf(
        "https://api.mapbox.com/directions/v5/mapbox/driving/%f,%f;%f,%f"+
            "?geometries=geojson"+
            "&access_token=%s",
        originLng, originLat, // Mapbox uses lng,lat order
        destLng, destLat,
        apiKey,
    )

    resp, err := http.Get(url)
    if err != nil {
        return 0, 0, 0, fmt.Errorf("mapbox API request failed: %v", err)
    }
    defer resp.Body.Close()

    var data DirectionsResponse
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return 0, 0, 0, fmt.Errorf("failed to parse Mapbox response: %v", err)
    }

    if data.Code != "Ok" {
        return 0, 0, 0, fmt.Errorf("mapbox API error: %s", data.Code)
    }

    if len(data.Routes) == 0 {
        return 0, 0, 0, fmt.Errorf("no routes found in Mapbox response")
    }

    // Convert to kilometers and minutes
    distance := data.Routes[0].Distance / 1000
    duration := data.Routes[0].Duration / 60

    surge, err := CalculateSurge()
    if err != nil {
        return 0, 0, 0, fmt.Errorf("failed to calculate surge: %v", err)
    }

    fare := CalculateFare(distance, vehicleType, surge)
    
    return distance, duration, fare, nil
}
