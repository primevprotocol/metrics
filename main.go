package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Define the request and response data structures
type RequestData struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      int      `json:"id"`
}

type ResponseData struct {
	Result struct {
		ExtraData string `json:"extraData"`
	} `json:"result"`
}

func main() {
	// Use zerolog for logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Initialize the params value (hexadecimal string)
	blockNumber := 0x10823a8

	// Define the URL
	url := "http://localhost:8545"

	for {
		// Prepare the request data
		requestData := RequestData{
			Jsonrpc: "2.0",
			Method:  "eth_getHeaderByNumber",
			Params:  []string{fmt.Sprintf("0x%x", blockNumber)},
			Id:      0,
		}

		jsonData, err := json.Marshal(requestData)
		if err != nil {
			log.Fatal().Err(err).Msg("Error marshalling JSON")
		}

		// Create the HTTP request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating HTTP request")
		}
		req.Header.Set("Content-Type", "application/json")

		// Send the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal().Err(err).Msg("Error sending HTTP request")
		}

		// Read the HTTP response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading HTTP response")
		}

		resp.Body.Close()

		// Decode the response JSON
		var responseData ResponseData
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding response JSON")
		}

		// Decode the extraData field from hex to a string
		extraDataBytes, err := hex.DecodeString(responseData.Result.ExtraData[2:]) // skip the '0x' prefix
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding extraData")
		}
		extraData := string(extraDataBytes)

		// Log the extraData
		log.Info().Int("block_number", blockNumber).Msg(extraData)

		// Increment the params value
		blockNumber++

		// Sleep for a while before starting the next loop
		time.Sleep(1 * time.Second)
	}
}
