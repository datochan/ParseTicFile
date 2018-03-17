package tic

type TickTradeDetail struct {
	Time		int   // 时间
	Price		int   // 价格
	Volume		int   // 成交
	Count		int   // 笔数
	Type		int   // 成交类型(0=buy;1=sell;2=unknown)
}

type StockTicModel struct {
	Market		int      // 沪深标识
	Code		string   // 股票代码
	Date		int      // 日期
	Details     []TickTradeDetail  // 分笔交易详情
}

type TickDetailModel struct {
	Date      int      // 日期
	Time      int      // 时间
	Price     int      // 价格
	Volume    int      // 成交
	Count     int      // 笔数
	Type      int      // 成交类型(0=buy;1=sell;2=unknown)
	VolOffset int      // 成交量的偏移地址
	VolSize   int
}
