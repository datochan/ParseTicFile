package main

import (
	"os"
	"ParseTicFile/tic"
	"fmt"
	"strings"
)

func main() {
	// parseticFile ./data/20180302.tic sz399001
	if len(os.Args) < 3 {
		fmt.Printf("Useage: ParseTicFile tickFilePath (sz|sh)stockCode \nexample: ParseTicFile ./data/20180302.tic sz000009")
		return
	}

	market := 0
	filePath := os.Args[1]
	strStockCode := os.Args[2]

	if strings.EqualFold(strStockCode[:2], "sh") {
		market = 1
	}

	tic.LoadTicFile(filePath, market, strStockCode[2:])

}
