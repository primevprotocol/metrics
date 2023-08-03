package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type ResponseData struct {
	Result struct {
		ExtraData string `json:"extraData"`
	} `json:"result"`
}
type Transaction struct {
	BlockHash            string `json:"blockHash"`
	BlockNumber          string `json:"blockNumber"`
	From                 string `json:"from"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
	Hash                 string `json:"hash"`
	Input                string `json:"input"`
	Nonce                string `json:"nonce"`
	To                   string `json:"to"`
	TransactionIndex     string `json:"transactionIndex"`
	Value                string `json:"value"`
	Type                 string `json:"type"`
	ChainId              string `json:"chainId"`
	V                    string `json:"v"`
	R                    string `json:"r"`
	S                    string `json:"s"`
}

type Withdrawal struct {
	Index          string `json:"index"`
	ValidatorIndex string `json:"validatorIndex"`
	Address        string `json:"address"`
	Amount         string `json:"amount"`
}

type Payload struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		BaseFeePerGas    string        `json:"baseFeePerGas"`
		Difficulty       string        `json:"difficulty"`
		ExtraData        string        `json:"extraData"`
		GasLimit         string        `json:"gasLimit"`
		GasUsed          string        `json:"gasUsed"`
		Hash             string        `json:"hash"`
		LogsBloom        string        `json:"logsBloom"`
		Miner            string        `json:"miner"`
		MixHash          string        `json:"mixHash"`
		Nonce            string        `json:"nonce"`
		Number           string        `json:"number"`
		ParentHash       string        `json:"parentHash"`
		ReceiptsRoot     string        `json:"receiptsRoot"`
		Sha3Uncles       string        `json:"sha3Uncles"`
		Size             string        `json:"size"`
		StateRoot        string        `json:"stateRoot"`
		Timestamp        string        `json:"timestamp"`
		TotalDifficulty  string        `json:"totalDifficulty"`
		Transactions     []Transaction `json:"transactions"`
		TransactionsRoot string        `json:"transactionsRoot"`
		Uncles           []string      `json:"uncles"`
		Withdrawals      []Withdrawal  `json:"withdrawals"`
		WithdrawalsRoot  string        `json:"withdrawalsRoot"`
	} `json:"result"`
}

type BlockInfo struct {
	Slot         int         `json:"slot"`
	Block        int         `json:"block"`
	BlockHash    string      `json:"block_hash"`
	Builder      string      `json:"builder"`
	Coinbase     string      `json:"coinbase"`
	Validator    string      `json:"validator"`
	FeeRecipient string      `json:"fee_recipient"`
	Payment      float64     `json:"payment"`
	Payout       float64     `json:"payout"`
	Payback      int         `json:"payback"`
	Alimony      float64     `json:"alimony"`
	Difference   float64     `json:"difference"`
	Offset       int         `json:"offset"`
	TxCount      int         `json:"tx_count"`
	GasUsed      int         `json:"gas_used"`
	BaseFee      float64     `json:"base_fee"`
	PrioFee      float64     `json:"prio_fee"`
	Extra        string      `json:"extra"`
	Winner       WinnerBlock `json:"winner"`
}

type WinnerBlock struct {
	Slot                 string          `json:"slot"`
	ParentHash           string          `json:"parent_hash"`
	BlockHash            string          `json:"block_hash"`
	BuilderPubkey        string          `json:"builder_pubkey"`
	ProposerPubkey       string          `json:"proposer_pubkey"`
	ProposerFeeRecipient string          `json:"proposer_fee_recipient"`
	GasLimit             string          `json:"gas_limit"`
	GasUsed              string          `json:"gas_used"`
	Value                string          `json:"value"`
	BlockNumber          string          `json:"block_number"`
	NumTx                string          `json:"num_tx"`
	Timestamp            string          `json:"timestamp"`
	TimestampMs          string          `json:"timestamp_ms"`
	OptimisticSubmission bool            `json:"optimistic_submission"`
	Builder              string          `json:"builder"`
	Relays               map[string]bool `json:"relays"`
}

func main() {
	// Use zerolog for logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Initialize the params value (hexadecimal string)
	blockNumberChan := make(chan int, 20)

	// Define the URL
	url := "http://localhost:8545"

	go func(blockNumbersChannel chan int) {
		blockNumber := 17815200

		for {
			requestData := RequestData{
				Jsonrpc: "2.0",
				Method:  "eth_blockNumber",
				Params:  []interface{}{},
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

			body, err := io.ReadAll(resp.Body)
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
		block := processBlockDataFromPayloadsDeAPI(blockNumber)
		log.Info().
			Int("block_number", block.Block).
			Int("txn_count", block.TxCount).
			Str("block_hash", block.BlockHash).
			Str("builder", block.Builder).
			Str("builder_pubkey", block.Winner.BuilderPubkey).
			Str("proposer_pubkey", block.Winner.ProposerPubkey).
			Float64("builder_payment", block.Payment).
			Float64("builder_payout", block.Payout).
			Str("block_value", block.Winner.Value).
			Str("extra_data", block.Extra).
			Float64("base_fee", block.BaseFee).
			Float64("priority_fee", block.PrioFee).
			Int("gas_used", block.GasUsed).
			Msg("New Block Metadata")
	}
}

func processBlockDataFromPayloadsDeAPI(blockNumber int) BlockInfo {
	// Make a request to the URL https://api.payload.de/block_info?block=17837129
	// to get the block information
	blockInfoURL := fmt.Sprintf("https://api.payload.de/block_info?block=%d", blockNumber)
	resp, err := http.Get(blockInfoURL)
	if err != nil {
		log.Error().Err(err).Msg("Error sending HTTP request")
		return BlockInfo{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading HTTP response")
		resp.Body.Close()
		return BlockInfo{}
	}
	resp.Body.Close()

	var blockInfo BlockInfo
	err = json.Unmarshal(body, &blockInfo)
	if err != nil {
		log.Error().Err(err).Msg("Error decoding response JSON")
		return BlockInfo{}
	}

	return blockInfo
}
