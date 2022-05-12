package tcp

import "github.com/pkg/errors"

var (
	// NoEnoughData 当前数据不够
	NoEnoughData = errors.New("no enough data")
	// BadMessageFormat 数据格式错
	BadMessageFormat = errors.New("no enough data")
)
