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
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      int      `json:"id"`
}

type BlockNumberResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"`
}

type ResponseData struct {
	Result struct {
		ExtraData string `json:"extraData"`
	} `json:"result"`
}

type ResponseDataBody struct {
	Result string `json:"result"`
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
				Params:  []string{},
				Id:      0,
			}

			jsonData, err := json.Marshal(requestData)
			if err != nil {
				log.Error().Err(err).Msg("Error marshalling JSON")
				time.Sleep(1 * time.Second)
				continue
			}

			req, err := http.NewRequest("POST", "http://localhost:8545", bytes.NewBuffer(jsonData))
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
		// Prepare the request data
		requestData := RequestData{
			Jsonrpc: "2.0",
			Method:  "eth_getHeaderByNumber",
			Params:  []string{fmt.Sprintf("0x%x", blockNumber)},
			Id:      0,
		}

		headerBody := processData(requestData, url)

		// Decode the response JSON
		var responseDataHeader ResponseData
		err := json.Unmarshal(headerBody, &responseDataHeader)
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding response JSON")
		}

		// Decode the extraData field from hex to a string
		extraDataBytes, err := hex.DecodeString(responseDataHeader.Result.ExtraData[2:]) // skip the '0x' prefix
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding extraData")
		}
		extraData := string(extraDataBytes)

		// Log the extraData
		log.Info().Int("block_number", blockNumber).Msg(extraData)

		// Prepare the request data for BlockByNumber
		requestDataBlockByNumber := RequestData{
			Jsonrpc: "2.0",
			Method:  "eth_getBlockByNumber",
			Params:  []string{fmt.Sprintf("0x%x", blockNumber), "true"},
			Id:      0,
		}

		responseDataBody := processData(requestDataBlockByNumber, url)

		// Decode the response JSON
		var responseBlock ResponseDataBody
		err = json.Unmarshal(responseDataBody, &responseBlock)
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding response JSON")
		}

		log.Info().Msg(responseBlock.Result)

		// Increment the params value
		blockNumber++
	}
}

func processData(requestData RequestData, url string) []byte {

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

	return body
}
