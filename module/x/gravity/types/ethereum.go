package types

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// GravityDenomPrefix indicates the prefix for all assests minted by this module
	GravityDenomPrefix = ModuleName

	// GravityDenomSeparator is the separator for gravity denoms
	GravityDenomSeparator = ""

	// ETHContractAddressLen is the length of contract address strings
	ETHContractAddressLen = 42

	// GravityDenomLen is the length of the denoms generated by the gravity module
	GravityDenomLen = len(GravityDenomPrefix) + len(GravityDenomSeparator) + ETHContractAddressLen
)

// EthAddrLessThan migrates the Ethereum address less than function
func EthAddrLessThan(e, o string) bool {
	return bytes.Compare([]byte(e)[:], []byte(o)[:]) == -1
}

// ValidateEthereumAddress validates the ethereum address strings
func ValidateEthereumAddress(a string) error {
	if a == "" {
		return fmt.Errorf("empty")
	}
	if !regexp.MustCompile("^0x[0-9a-fA-F]{40}$").MatchString(a) {
		return fmt.Errorf("address(%s) doesn't pass regex", a)
	}
	if len(a) != ETHContractAddressLen {
		return fmt.Errorf("address(%s) of the wrong length exp(%d) actual(%d)", a, len(a), ETHContractAddressLen)
	}
	return nil
}

/////////////////////////
//     ERC20Token      //
/////////////////////////

// NewERC20Token returns a new instance of an ERC20
func NewERC20Token(amount uint64, contract string) ERC20Token {
	return ERC20Token{
		Amount:   sdk.NewIntFromUint64(amount),
		Contract: contract,
	}
}

func NewSDKIntERC20Token(amount sdk.Int, contract common.Address) ERC20Token {
	return ERC20Token{
		Amount:   amount,
		Contract: contract.Hex(),
	}
}

// GravityCoin returns the gravity representation of the ERC20
func (e ERC20Token) GravityCoin() sdk.Coin {
	return sdk.Coin{Amount: e.Amount, Denom: strings.Join([]string{GravityDenomPrefix, e.Contract}, GravityDenomSeparator)}
}

func NewERC20TokenFromCoin(coin sdk.Coin) ERC20Token {
	return NewSDKIntERC20Token(coin.Amount, common.HexToAddress(strings.TrimPrefix(coin.Denom, GravityDenomPrefix+GravityDenomSeparator)))
}

// ValidateBasic permforms stateless validation
func (e *ERC20Token) ValidateBasic() error {
	if err := ValidateEthereumAddress(e.Contract); err != nil {
		return sdkerrors.Wrap(err, "ethereum address")
	}
	// TODO: Validate all the things
	return nil
}

// Add adds one ERC20 to another
// TODO: make this return errors instead
func (e *ERC20Token) Add(o ERC20Token) ERC20Token {
	if e.Contract != o.Contract {
		panic("invalid contract address")
	}
	sum := e.Amount.Add(o.Amount)
	if !sum.IsUint64() {
		panic("invalid amount")
	}
	return NewERC20Token(sum.Uint64(), e.Contract)
}

func GravityDenomToERC20(denom string) (string, error) {
	fullPrefix := GravityDenomPrefix + GravityDenomSeparator
	if !strings.HasPrefix(denom, fullPrefix) {
		return "", fmt.Errorf("denom prefix(%s) not equal to expected(%s)", denom, fullPrefix)
	}
	contract := strings.TrimPrefix(denom, fullPrefix)
	err := ValidateEthereumAddress(contract)
	switch {
	case err != nil:
		return "", fmt.Errorf("error(%s) validating ethereum contract address", err)
	case len(denom) != GravityDenomLen:
		return "", fmt.Errorf("len(denom)(%d) not equal to GravityDenomLen(%d)", len(denom), GravityDenomLen)
	default:
		return contract, nil
	}
}

func NewSendToEthereumTx(id uint64, tokenContract common.Address, sender sdk.AccAddress, recipient common.Address, amount, feeAmount uint64) *SendToEthereum {
	return &SendToEthereum{
		Id:                id,
		Erc20Fee:          NewERC20Token(feeAmount, tokenContract.Hex()),
		Sender:            sender.String(),
		EthereumRecipient: recipient.Hex(),
		Erc20Token:        NewERC20Token(amount, tokenContract.Hex()),
	}
}

// Id:                2,
// Erc20Fee:          types.NewERC20Token(3, myTokenContractAddr),
// Sender:            mySender.String(),
// EthereumRecipient: myReceiver,
// Erc20Token:        types.NewERC20Token(101, myTokenContractAddr),
