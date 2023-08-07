package main

import (
	"encoding/json"
	"errors"
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

type BlockInfo struct {
	Slot         int           `json:"slot"`
	Block        int           `json:"block"`
	BlockHash    string        `json:"block_hash"`
	Builder      string        `json:"builder"`
	Coinbase     string        `json:"coinbase"`
	Validator    string        `json:"validator"`
	FeeRecipient string        `json:"fee_recipient"`
	Payment      float64       `json:"payment"`
	Payout       float64       `json:"payout"`
	Payback      float64       `json:"payback"`
	Alimony      float64       `json:"alimony"`
	Difference   float64       `json:"difference"`
	Offset       int           `json:"offset"`
	TxCount      int           `json:"tx_count"`
	GasUsed      int           `json:"gas_used"`
	BaseFee      float64       `json:"base_fee"`
	PrioFee      float64       `json:"prio_fee"`
	Extra        string        `json:"extra"`
	Timestamp    int           `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	Winner       WinnerBlock   `json:"winner"`
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

type Transaction struct {
	Hash        string   `json:"hash"`
	Block       int      `json:"block"`
	Class       string   `json:"class"`
	MethodID    string   `json:"method_id"`
	Method      string   `json:"method"`
	TargetLabel string   `json:"target_label"`
	Tags        []string `json:"tags"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	DataSize    int      `json:"data_size"`
	GasLimit    int      `json:"gas_limit"`
	GasUsed     int      `json:"gas_used"`
	Status      int      `json:"status"`
	Fee         float64  `json:"fee"`
	Tip         float64  `json:"tip"`
	Price       float64  `json:"price"`
	TipReward   float64  `json:"tip_reward"`
	Value       float64  `json:"value"`
	Nonce       int      `json:"nonce"`
	Position    int      `json:"position"`
	Coinbase    float64  `json:"coinbase"`
	Timestamp   int      `json:"timestamp"`
}

func main() {
	// Use zerolog for logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimestampFieldName = "printout_time"
	for blockNumber := 17864000; ; blockNumber++ {
		block, err := processBlockDataFromPayloadsDeAPI(blockNumber)
		if err == Err400Response {
			time.Sleep(12 * time.Second)
			blockNumber--
			continue
		} else if err != nil {
			log.Error().Int("block_number", blockNumber).Err(err).Msg("Error processing block data, skipping")
			continue
		}

		floatBlockValue, err := strconv.ParseFloat(block.Winner.Value, 64)
		if err != nil {
			log.Error().Err(err).Msg("Error converting block value to float")
		}
		categoryCounter := make(map[string]int)
		valueCounter := make(map[string]float64)
		for _, txn := range block.Transactions {
			categoryCounter[txn.Class]++
			valueCounter[txn.Class] += txn.Value
		}
		mevCount := float64(categoryCounter["mev"]) / float64(block.TxCount)
		mevValue := valueCounter["mev"] / floatBlockValue

		log.Info().
			Str("log_version", "1.4").
			Int("block_number", block.Block).
			Int("txn_count", block.TxCount).
			Str("block_hash", block.BlockHash).
			Str("builder", block.Builder).
			Str("builder_pubkey", block.Winner.BuilderPubkey).
			Str("builder_address", block.Winner.BuilderPubkey).
			Str("proposer_pubkey", block.Winner.ProposerPubkey).
			Float64("builder_payment", block.Payment).
			Float64("builder_payout", block.Payout).
			Float64("block_value_priority_fee", floatBlockValue).
			Str("extra_data", block.Extra).
			Float64("base_fee", block.BaseFee).
			Float64("priority_fee", block.PrioFee).
			Float64("txn_count_mev_percentage", mevCount*100).
			Float64("txn_value_mev_percentage", mevValue*100).
			Int("gas_used", block.GasUsed).
			Str("time", time.Unix(int64(block.Timestamp), 0).Format(time.RFC3339)).
			Msg("New Block Metadata V1.4")
	}
}

var Err400Response = errors.New("received 400 response from payloads.de")
var ErrUnableToUnmarshal = errors.New("unable to unmarshal response JSON")

// Check if request returns a 400 and wait 12 seconds before trying again. It's likeley payloads.de hasn't caught up witht he new block yet
func processBlockDataFromPayloadsDeAPI(blockNumber int) (BlockInfo, error) {
	// Make a request to the URL https://api.payload.de/block_info?block=17837129
	// to get the block information
	time.Sleep(500 * time.Nanosecond)
	blockInfoURL := fmt.Sprintf("https://api.payload.de/block_info?block=%d", blockNumber)
	resp, err := http.Get(blockInfoURL)
	if err != nil {
		log.Error().Err(err).Msg("Error sending HTTP request")
		return BlockInfo{}, err
	}
	if resp.StatusCode == 400 {
		log.Warn().Msgf("Received 400 response from payloads.de. Waiting 12 seconds before trying again")

		return BlockInfo{}, Err400Response
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading HTTP response")
		resp.Body.Close()
		return BlockInfo{}, Err400Response
	}
	resp.Body.Close()

	var blockInfo BlockInfo
	err = json.Unmarshal(body, &blockInfo)
	if err != nil {
		log.Error().Err(err).Msg("Error decoding response JSON")
		return BlockInfo{}, ErrUnableToUnmarshal
	}

	return blockInfo, nil
}
