package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Define the request and response data structures
type RequestData struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type BlockNumberResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"`
}

type BlockDataResponse struct {
	Result struct {
		ExtraData string `json:"extraData"`
	} `json:"result"`
}

func processRequest(requestData RequestData, blockNumber int) string {
	url := "http://localhost:8545"

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		log.Fatal().Err(err).Msg("Error marshalling JSON")
		return ""
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating HTTP request")
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("Error sending HTTP request")
		return ""
	}

	// Read the HTTP response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading HTTP response")
		return ""
	}

	resp.Body.Close()

	// Decode the response JSON
	var blockDataResponse BlockDataResponse
	err = json.Unmarshal(body, &blockDataResponse)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding response JSON")
		return ""
	}

	// Decode the extraData field from hex to a string
	extraDataBytes, err := hex.DecodeString(blockDataResponse.Result.ExtraData[2:]) // skip the '0x' prefix
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding extraData")
		return ""
	}
	extraData := string(extraDataBytes)

	return extraData
}

func main() {
	// Use zerolog for logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Initialize the params value (hexadecimal string)
	blockNumberChan := make(chan int, 20)

	// Define the URL
	url := "http://localhost:8545"

	go func(blockNumbersChannel chan int) {
		blockNumber := 17785600

		for {
			requestData := RequestData{
				Jsonrpc: "2.0",
				Method:  "eth_blockNumber",
				Params:  nil,
				Id:      0,
			}

			jsonData, err := json.Marshal(requestData)
			if err != nil {
				log.Error().Err(err).Msg("Error marshalling JSON")
				time.Sleep(1 * time.Second)
				continue
			}

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Error().Err(err).Msg("Error creating HTTP request")
				time.Sleep(1 * time.Second)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Error().Err(err).Msg("Error sending HTTP request")
				time.Sleep(1 * time.Second)
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error().Err(err).Msg("Error reading HTTP response")
				resp.Body.Close()
				time.Sleep(1 * time.Second)
				continue
			}
			resp.Body.Close()

			var blockNumberResponse BlockNumberResponse
			err = json.Unmarshal(body, &blockNumberResponse)
			if err != nil {
				log.Error().Err(err).Msg("Error decoding response JSON")
				time.Sleep(1 * time.Second)
				continue
			}

			newBlockNumber, err := strconv.ParseUint(blockNumberResponse.Result[2:], 16, 64)
			if err != nil {
				log.Error().Err(err).Msgf("Error converting block number %s to decimal", blockNumberResponse.Result)
				continue
			}
			for blockNumber < int(newBlockNumber) {
				blockNumber = blockNumber + 1
				blockNumbersChannel <- blockNumber
			}
			time.Sleep(1 * time.Second)
		}
	}(blockNumberChan)

	for blockNumber := range blockNumberChan {
		// Prepare the request data for BlockByNumber
		requestDataBlockByNumber := RequestData{
			Jsonrpc: "2.0",
			Method:  "eth_getBlockByNumber",
			Params:  []interface{}{fmt.Sprintf("0x%x", blockNumber), true},
			Id:      0,
		}

		// Prepare the request data for HeaderByNumber
		requestDataHeaderByNumber := RequestData{
			Jsonrpc: "2.0",
			Method:  "eth_getHeaderByNumber",
			Params:  []interface{}{fmt.Sprintf("0x%x", blockNumber), true},
			Id:      0,
		}

		extraDataBlockByNumber := processRequest(requestDataBlockByNumber, blockNumber)
		extraDataHeaderByNumber := processRequest(requestDataHeaderByNumber, blockNumber)

		// Log the extraData
		log.Info().Int("block_number", blockNumber).Msgf("BlockByNumber: %s, HeaderByNumber: %s", extraDataBlockByNumber, extraDataHeaderByNumber)

		// Increment the params value
		blockNumber++
	}
}
