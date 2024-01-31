package main

import (
	c "Go-API-Fetch/config"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

func printCSV(file *os.File) error {
	if file == nil {
		return fmt.Errorf("file is nil")
	}

	var numRows = 0
	reader := csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		fmt.Println(record)
		numRows++
	}

	return nil
}

type BerryFirmnessResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}
type BerryData struct {
	Berries []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"berries"`
}

func main() {

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

	file, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"name", "url"})

	totalPages := 1

	// Send request to API until all data is fetched
	for page := 0; page <= totalPages; page++ {
		fmt.Print("Fetching page ", page, " of ", totalPages, "\n")
		var fullEndPoint = configuration.Api.EndPoint + configuration.Api.Operation
		if page > 0 {
			fullEndPoint += strconv.Itoa(page)
		}

		fmt.Print("URL: ", fullEndPoint, "\n")

		req, err := http.NewRequest("GET", fullEndPoint, nil)

		if err != nil {
			panic(err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("charset", "UTF-8")
		req.Header.Set("Authorization", "serviceKey")
		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		if page == 0 {
			// Unmarshal JSON data into struct
			var berryFirmness BerryFirmnessResponse
			err = json.Unmarshal(body, &berryFirmness)
			if err != nil {
				panic(err)
			}
			print(berryFirmness.Results)
			totalPages = berryFirmness.Count
		} else {
			var berryData BerryData
			if err := json.Unmarshal(body, &berryData); err != nil {
				panic(err)
			}

			// Write berry data to CSV
			for _, item := range berryData.Berries {
				if err := writer.Write([]string{item.Name, item.URL}); err != nil {
					panic(err)
				}
			}
		}

	}
	file, err = os.Open("output.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer file.Close()

	// Call printCSV to print the file
	err = printCSV(file)
	if err != nil {
		fmt.Println("Error:", err)
	}

}
