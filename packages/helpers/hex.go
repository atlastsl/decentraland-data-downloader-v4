package helpers

import (
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func HexRemoveLeadingZerosTrimLeft(s string) string {
	tmp := strings.TrimLeft(s, "0")
	if tmp == "" {
		tmp = "0"
	}
	return tmp
}

func HexRemoveLeadingZeros(s string) string {
	if strings.HasPrefix(s, "0x") {
		return "0x" + HexRemoveLeadingZerosTrimLeft(s[2:])
	}
	return HexRemoveLeadingZerosTrimLeft(s)
}

func HexConvertToInt(s string) (int, error) {
	cleanHex := HexRemoveLeadingZeros(s)
	bs, err := hexutil.DecodeUint64(cleanHex)
	if err != nil {
		return 0, err
	}
	return int(bs), nil
}

func HexConvertToString(s string) (string, error) {
	cleanHex := HexRemoveLeadingZeros(s)
	bs, err := hexutil.DecodeBig(cleanHex)
	if err != nil {
		return "", err
	}
	return bs.String(), nil
}

func HexConvertToFloat64(s string) (float64, error) {
	cleanHex := HexRemoveLeadingZeros(s)
	bs, err := hexutil.DecodeBig(cleanHex)
	if err != nil {
		return 0.0, err
	}
	bf, _ := bs.Float64()
	return bf, nil
}
