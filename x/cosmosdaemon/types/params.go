package types

import (
	"errors"
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	DefaultDenom    string = "UGD"
	DefaultAmount   uint64 = 100000000
	DefaultMaxRate  uint64 = 9999
	DefaultRate     uint64 = 888
	DefaultMaxBytes uint64 = 77777
	DefaultMaxGas   uint64 = 666
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(denom string, amount uint64, maxRate uint64, rate uint64, maxBytes uint64, maxGas uint64) Params {
	return Params{
		Denom:    denom,
		Amount:   amount,
		MaxRate:  maxRate,
		Rate:     rate,
		MaxBytes: maxBytes,
		MaxGas:   maxGas,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultDenom,
		DefaultAmount,
		DefaultMaxRate,
		DefaultRate,
		DefaultMaxBytes,
		DefaultMaxGas,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDenom(p.Denom); err != nil {
		return err
	}

	if err := validateAmount(p.Amount); err != nil {
		return err
	}

	if err := validateMaxRate(p.MaxRate); err != nil {
		return err
	}

	if err := validateRate(p.Rate); err != nil {
		return err
	}

	if err := validateMaxBytes(p.MaxBytes); err != nil {
		return err
	}

	if err := validateMaxGas(p.MaxGas); err != nil {
		return err
	}

	return nil
}

func validateAmount(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("amount must be positive: %d", v)
	}

	return nil
}

func validateDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("denom cannot be blank")
	}

	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateMaxRate(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("maxRate must be positive: %d", v)
	}

	return nil
}

func validateRate(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("rate must be positive: %d", v)
	}

	return nil
}

func validateMaxBytes(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("maxBytes must be positive: %d", v)
	}

	return nil
}

func validateMaxGas(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("maxGas must be positive: %d", v)
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
