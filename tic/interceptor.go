package tic

import (
	"fmt"
	"math"
	"bytes"
	"io/ioutil"
	"runtime/debug"
	"encoding/binary"

	gbytes "github.com/datochan/gcom/bytes"
	"github.com/datochan/gcom/utils"
)

/**
 * timeVal: 开市后经历的分钟数
 * return: 调整为当天的分钟数
 */
func SetTradeTime(timeVal int) string {
	result := 0x23A
	if timeVal >= 0 && timeVal <= 0x78 {
		result += timeVal
	} else if timeVal <= 0xF0 {
		result = 0x294 + timeVal
	}

	return fmt.Sprintf("%02d:%02d", int(math.Floor(float64(result/60%24))), result%60)
}

func parseTickDTPrice(tradeDetails []TickTradeDetail, leBuffer *gbytes.LittleEndianStreamImpl, tickItem TickDetailModel) ([]TickTradeDetail, error) {
	tmpSize := 32

	// 0x949D70AA
	// 1001 0100 1001 1101 0111 0000 1010 1010
	// -  type
	detailOffset := leBuffer.Size()-leBuffer.Right()

	tickDataItem, err := leBuffer.ReadUint32()
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	for idx:=1; idx<tickItem.Count; idx ++{
		var tradeDetail TickTradeDetail

		tradeDetail.Type = int(tickDataItem >> 31)
		tickDataItem <<= 1
		tmpSize --
		if tmpSize == 0 {
			if detailOffset >= int(tickItem.VolOffset+tickItem.VolSize+0x10) {
				return nil, fmt.Errorf("tic文件解析失败: 偏移量超出额定范围")
			}

			tickDataItem, err = leBuffer.ReadUint32()
			if err != nil {
				debug.PrintStack()
				return nil, err
			}

			detailOffset = leBuffer.Size()-leBuffer.Right()
			tmpSize = 32
		}

		tmpCheckSum := uint32(3)

	LABEL1:
		tmpCheckSum = (2 * tmpCheckSum) | (tickDataItem >> 31)
		tickDataItem <<= 1
		tmpSize --

		if tmpSize == 0 {
			tickDataItem, err = leBuffer.ReadUint32()
			detailOffset = leBuffer.Size()-leBuffer.Right()
			if err != nil {
				fmt.Printf("读取数据时发生错误:%s \n", err.Error())
				debug.PrintStack()
				return nil, err
			}

			tmpSize = 32
		}

		tmpIdx := 0
		for HashTableDateTime[tmpIdx].HashValue != int32(tmpCheckSum) {
			if HashTableDateTime[tmpIdx].HashValue < int32(tmpCheckSum) {
				tmpIdx++
				if tmpIdx < len(HashTableDateTime){
					continue
				}
			}

			goto LABEL1
		}

		// 解析时间
		tradeDetail.Time = int(tradeDetails[len(tradeDetails)-1].Time) + HashTableDateTime[tmpIdx].Idx

		tmpCheckSum = 3
	LABEL2:
		tmpCheckSum = (2 * tmpCheckSum) | (tickDataItem >> 31)
		tickDataItem <<= 1
		tmpSize --

		if tmpSize == 0 {
			tickDataItem, err = leBuffer.ReadUint32()
			detailOffset = leBuffer.Size()-leBuffer.Right()
			if err != nil {
				fmt.Printf("读取数据时发生错误:%s \n", err.Error())
				debug.PrintStack()
				return nil, err
			}

			tmpSize = 32
		}

		tmpIdx = 0
		for HashTablePrice[tmpIdx].HashValue != int32(tmpCheckSum) {
			if tmpCheckSum > 0x3FFFFFF || HashTablePrice[tmpIdx].HashValue <= int32(tmpCheckSum) {
				tmpIdx++
				if tmpIdx < len(HashTablePrice){
					continue
				}
			}

			goto LABEL2
		}

		if tmpIdx != 4000 || tickItem.Date < 20011112 {
			tradeDetail.Price = int(tradeDetails[len(tradeDetails)-1].Price) + HashTablePrice[tmpIdx].Idx
		} else {
			tmpCheckSum = 0

			for tmpIdx = 32; tmpIdx > 0; tmpIdx -- {
				tmpCheckSum = (2 * tmpCheckSum) | (tickDataItem >> 31)
				tickDataItem <<= 1
				tmpSize --
				if tmpSize == 0 {
					tickDataItem, err = leBuffer.ReadUint32()
					detailOffset = leBuffer.Size()-leBuffer.Right()
					if err != nil {
						fmt.Printf("读取数据时发生错误:%s \n", err.Error())
						debug.PrintStack()
						return nil, err
					}

					tmpSize = 32
				}
			}

			tradeDetail.Price = int(tradeDetails[len(tradeDetails)-1].Price) + int(tmpCheckSum)
		}

		tradeDetails = append(tradeDetails, tradeDetail)
	}

	return tradeDetails, nil
}

func parseTickDetail(byTicDetail []byte, tickItem TickDetailModel) ([]TickTradeDetail, error) {

	var err error
	if tickItem.Count <= 0 {
		return nil, fmt.Errorf("tic文件解析失败: 数量解析失败")
	}

	leBuffer := gbytes.NewLittleEndianStream(byTicDetail)
	var tradeDetails []TickTradeDetail
	tradeDetails = append(tradeDetails, TickTradeDetail{tickItem.Time, tickItem.Price, tickItem.Volume,
		tickItem.Count, tickItem.Type >> 31})

	// 解析交易时间及价格信息
	tradeDetails, err = parseTickDTPrice(tradeDetails, leBuffer, tickItem)
	if err != nil {
		return nil, err
	}

	// 解析成交量
	volumeBuffer := gbytes.NewLittleEndianStream(byTicDetail[tickItem.VolOffset:])

	for idx := 1; idx < tickItem.Count; idx ++ {
		resultVol := 0
		byteVolume, err := volumeBuffer.ReadByte()
		if err != nil {
			debug.PrintStack()
			return nil, err
		}

		if byteVolume <= 252 {
			resultVol = int(byteVolume)
		} else if byteVolume == 253 {
			tmpVol, _ := volumeBuffer.ReadByte()
			resultVol = int(tmpVol+byteVolume)

		} else if byteVolume == 254 {
			tmpVol, _ := volumeBuffer.ReadUint16()
			resultVol = int(tmpVol+uint16(byteVolume))

		} else if byteVolume == 255 {
			tmpVol1, _ := volumeBuffer.ReadByte()
			tmpVol2, _ := volumeBuffer.ReadUint16()
			resultVol = int(0xFFFF * int(tmpVol1) + int(tmpVol2) + 0xFF)
		}

		tradeDetails[idx].Volume = resultVol
	}

	return tradeDetails, nil
}

func ParseTickItem(byteTic []byte) {
	var tickItem TickItem
	var newBuffer bytes.Buffer
	var tickDetailModel TickDetailModel

	leBuffer := gbytes.NewLittleEndianStream(byteTic)

	rawTickItem, _ := leBuffer.ReadBuff(utils.SizeStruct(TickItem{}))
	newBuffer.Write(rawTickItem)
	binary.Read(&newBuffer, binary.LittleEndian, &tickItem)

	tickDetailModel.Date = int(tickItem.DateTime)
	tickDetailModel.Time = int(byte(tickItem.Type))
	tickDetailModel.Price = int(tickItem.Price)
	tickDetailModel.Volume = int(tickItem.Volume)
	tickDetailModel.Count = int(tickItem.Count)
	tickDetailModel.Type = int(tickItem.Type)
	tickDetailModel.VolOffset = int(tickItem.VolOffset)
	tickDetailModel.VolSize = int(tickItem.VolSize)

	byteTicDetail, _ := leBuffer.ReadBuff(leBuffer.Right())

	tradeDetails, err := parseTickDetail(byteTicDetail, tickDetailModel)

	if nil != err {
		fmt.Printf("解析tck详情时报错: %s", err.Error())
		return
	}

	fmt.Printf("\t时间,\t价格,\t交易量,\t笔数,\t交易方向\n")
	for _, item := range tradeDetails {
		fmt.Printf("\t%s,\t%.2f,\t%d,\t%d,\t%d\n", SetTradeTime(item.Time),
			float64(item.Price)/100.0, item.Volume, item.Count, item.Type)
	}
}

func LoadTicFile(filePath string, market int, stockCode string) error {
	var newBuffer bytes.Buffer
	var stockTick StockTick
	byteTic, err := ioutil.ReadFile(filePath)
	if nil != err {
		return err
	}

	leBuffer := gbytes.NewLittleEndianStream(byteTic)

	stockCount, _ := leBuffer.ReadUint16()
	fmt.Printf("股票数量为: %d\n", stockCount)

	for idx:=0; idx < int(stockCount); idx++{
		rawStockTick, _ := leBuffer.ReadBuff(utils.SizeStruct(StockTick{}))
		newBuffer.Write(rawStockTick)

		binary.Read(&newBuffer, binary.LittleEndian, &stockTick)

		tickSize := int(stockTick.TickSize)
		rawTickData, _ := leBuffer.ReadBuff(tickSize)

		strCode := gbytes.BytesToString(stockTick.Code[:])

		if int(stockTick.Market) == market && strCode == stockCode {
			fmt.Printf("开始解析股票: %d%s, date: %d\n", stockTick.Market,
				gbytes.BytesToString(stockTick.Code[:]), stockTick.Date)
			ParseTickItem(rawTickData)
			break
		}

	}

	return nil
}