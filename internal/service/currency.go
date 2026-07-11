package service

import (
	"github.com/afadhitya/wallet-app/internal/service/shared"
)

func SetTestRateConfig(cfg shared.TestRateConfig) {
	shared.SetTestRateConfig(cfg)
}

func ResetTestRateConfig() {
	shared.ResetTestRateConfig()
}

type RateNotFoundError = shared.RateNotFoundError

type RateInfo = shared.RateInfo

var (
	ErrRateConfigMissing  = shared.ErrRateConfigMissing
	ErrRateMustBePositive = shared.ErrRateMustBePositive
)
