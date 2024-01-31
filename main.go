package main

import (
	c "Go-API-Fetch/config"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type BerryFirmnessResponse struct {
	Count int `json:"count"`
}

// Initialize Viper to read configuration
func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	var configuration c.Configurations

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Set undefined variables
	viper.SetDefault("database.dbname", "test_db")
	err := viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

}

func fetchTotalItems() (int, error) {
	baseURL := viper.GetString("api.endPoint")
	endpoint := viper.GetString("api.operation")
	fullURL := baseURL + endpoint

	resp, err := http.Get(fullURL)
	if err != nil {
		return 0, fmt.Errorf("error fetching total items: %w", err)
	}
	defer resp.Body.Close()

	var response BerryFirmnessResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Count, nil
}

func fetchItemDetails(itemID int, wg *sync.WaitGroup, detailsChan chan<- string) {
	defer wg.Done()
	baseURL := viper.GetString("api.endPoint") + viper.GetString("api.operation")
	itemURL := fmt.Sprintf("%s%d", baseURL, itemID)

	resp, err := http.Get(itemURL)
	if err != nil {
		fmt.Printf("Failed to fetch item %d: %v\n", itemID, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response for item %d: %v\n", itemID, err)
		return
	}

	detailsChan <- string(body)
}

func main() {
	initConfig() // Initialize configuration

	totalItems, err := fetchTotalItems()
	if err != nil {
		fmt.Printf("Error fetching total items: %v\n", err)
		return
	}

	fmt.Printf("Total Items: %d\n", totalItems)

	detailsChan := make(chan string, totalItems)
	var wg sync.WaitGroup

	for i := 1; i <= totalItems; i++ {
		wg.Add(1)
		go fetchItemDetails(i, &wg, detailsChan)
	}

	go func() {
		wg.Wait()
		close(detailsChan)
	}()

	file, err := os.Create("berry_details.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for detail := range detailsChan {
		if err := writer.Write([]string{detail}); err != nil {
			fmt.Println("Error writing to CSV:", err)
			return
		}
	}

	fmt.Println("Finished writing item details to CSV.")
}
