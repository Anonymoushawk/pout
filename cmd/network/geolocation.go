package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// `GeoLocation` represents the information about a geographical location derived from an IP Address.
type GeoLocation struct {
	Country    string `json:"country"`
	RegionName string `json:"region_name"`
	City       string `json:"city"`
	ZipCode    string `json:"zip_code"`
	ASNumber   string `json:"as_number"`
}

// `GetGeoLocation` retrieves the geolocation information for the given IP address using the GEO_IP_API.
func GetGeoLocation(ip string) (GeoLocation, error) {
	var geoData = GeoLocation{
		// Set the default GeoLocation values using the DEFAULT_GEO_VAL constant.
		Country:    DEFAULT_GEO_VAL,
		RegionName: DEFAULT_GEO_VAL,
		City:       DEFAULT_GEO_VAL,
		ZipCode:    DEFAULT_GEO_VAL,
		ASNumber:   DEFAULT_GEO_VAL,
	}

	// Format the GEO_IP_API to include the passed ip address for reading.
	url := fmt.Sprintf("%s/%s", GEO_IP_API, ip)

	// Create a GET request with a mock User-Agent that we can send to the GEO_IP_API.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return geoData, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Send the aforementioned GET request to the GEO_IP_API.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return geoData, err
	}
	defer resp.Body.Close()

	// Read the requests response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return geoData, err
	}

	// Read the JSON request body into a new data map.
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return geoData, err
	}

	// Check if the expected keys exist in the data map.
	// If the GEO_IP_API cannot be reached, or there is an error encountered
	// when reading the data map, we will be able to check for that here, therefore
	// avoiding any JSON read errors or exception panics, that could possibly shut down the server.
	if country, ok := data["country"].(string); ok {
		geoData.Country = country
	}
	if regionName, ok := data["regionName"].(string); ok {
		geoData.RegionName = regionName
	}
	if city, ok := data["city"].(string); ok {
		geoData.City = city
	}
	if zip, ok := data["zip"].(string); ok {
		geoData.ZipCode = zip
	}
	if asNumber, ok := data["as"].(string); ok {
		geoData.ASNumber = asNumber
	}

	return geoData, nil
}

const (
	// IP GeoLocation API.
	GEO_IP_API = "http://ip-api.com/json"

	// Default value to use in the GeoLocation struct.
	// This is used when the GEO_IP_API cannot be reached.
	DEFAULT_GEO_VAL = "Unknown"
)
