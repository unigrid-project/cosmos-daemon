package types

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ErrorWhenGettingCache struct{}

func (e *ErrorWhenGettingCache) Error() string {
	return "Failed to get address from cache, cache is probably empty"
}

type Vesting struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type ConsensusBlock struct {
	MaxBytes string `json:"maxBytes"`
	MaxGas   string `json:"maxGas"`
}

type GenesisTransactions struct {
	Rate    string `json:"rate"`
	MaxRate string `json:"maxRate"`
}

type UgdParam struct {
	Vesting             `json:"vesting"`
	ConsensusBlock      `json:"consensusBlock"`
	GenesisTransactions `json:"genesisTransactions"`
}

type Data struct {
	Parameters UgdParam `json:"parameters"`
}

type HedgehogData struct {
	TimeStamp         time.Time `json:"timeStamp"`
	PreviousTimeStamp time.Time `json:"previousTimeStamp"`
	Data              Data      `json:"data"`
}

type ParamCache struct {
	stop chan struct{}

	wg     sync.WaitGroup
	mu     sync.RWMutex
	params map[string]UgdParam
}

const (
	cacheUpdateInterval = 1 * time.Minute
)

var pcg = NewCache()

func (pc *ParamCache) cleanupCache() {
	t := time.NewTicker(cacheUpdateInterval)
	defer t.Stop()

	for {
		select {
		case <-pc.stop:
			return
		case <-t.C:
			pc.mu.Lock()
			pc.callHedgehog("https://127.0.0.1:52884/gridspork/cosmos")
			pc.mu.Unlock()
		}
	}
}

func GetCache() *ParamCache {
	fmt.Println("Getting cache")

	if pcg == nil {
		pcg = NewCache()
	}
	return pcg
}

func NewCache() *ParamCache {
	pc := &ParamCache{
		params: make(map[string]UgdParam),
		stop:   make(chan struct{}),
	}

	pc.wg.Add(1)
	go func() {
		defer pc.wg.Done()
		pc.cleanupCache()
	}()

	return pc
}

func (pc *ParamCache) GetUgdParams() (UgdParam, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	ugdp, ok := pc.params["parameters"]
	if !ok {
		return UgdParam{}, &ErrorWhenGettingCache{}
	}

	return ugdp, nil
}

func StringToUint(str string) uint64 {
	ui64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		panic(err)
	}

	return ui64
}

func (pc *ParamCache) callHedgehog(serverURl string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Get(serverURl)

	if err != nil {
		panic("where is hedgehog " + err.Error())
	}

	defer response.Body.Close()
	var res HedgehogData
	body, err1 := io.ReadAll(response.Body)

	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}

	e := json.Unmarshal(body, &res)
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	pc.params["parameters"] = res.Data.Parameters

	// fmt.Println("RESPONSE----------------------------------->")
	// fmt.Println("Timestamp:", res.TimeStamp)
	// fmt.Println("Previous Timestamp:", res.PreviousTimeStamp)
	// fmt.Println("Vesting Denom:", res.Data.Parameters.Vesting.Denom)
	// fmt.Println("Vesting Amount:", res.Data.Parameters.Vesting.Amount)
	// fmt.Println("Genesis Transactions Rate:", res.Data.Parameters.GenesisTransactions.Rate)
	// fmt.Println("Genesis Transactions MaxRate:", res.Data.Parameters.GenesisTransactions.MaxRate)
	// fmt.Println("Consensus Block MaxBytes:", res.Data.Parameters.ConsensusBlock.MaxBytes)
	// fmt.Println("Consensus Block MaxGas:", res.Data.Parameters.ConsensusBlock.MaxGas)
}
