package utils

import (
	"os"
	"sync"
)

var (
	pricesFilePath     string
	pricesFilePathOnce sync.Once
)

func ResetPricesFilePathCacheForTesting() {
	pricesFilePath = ""
	pricesFilePathOnce = sync.Once{}
}

func GetPricesFilePath() string {
	pricesFilePathOnce.Do(func() {
		filePath := "internal/data/prices.json"
		if _, err := os.Stat(filePath); err != nil {
			if _, err2 := os.Stat("../../internal/data/prices.json"); err2 == nil {
				filePath = "../../internal/data/prices.json"
			}
		}
		pricesFilePath = filePath
	})
	return pricesFilePath
}
