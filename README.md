# ParseTicFile
通达信tic文件格式解析

代码是直接逆向分析出来的，没有优化(比较挫,但愿大家能看懂!)，只用来说明Tic数据的文件格式。

数据的准确性我没有认真校对，应该差不多，如下图:
![示例图片](https://raw.githubusercontent.com/datochan/ParseTicFile/master/data/example.png)

# 命令的使用方法

```$bash
Useage:
    parseticFile TickFilePath (sz|sh)stockCode
    
example:
    ParseTicFile ./data/20180302.tic sz000009
```

# Tic文件的获取方法

分笔数据的获取方法:

* 下载索引文件: `http://www.tdx.com.cn/downit.zip`
* 解析这个压缩包中的配置文件`downit5.cfg` 
> 这里面有基本面数据权息数据、日线数据、分笔数据等各种数据的下载url。
* `/products/data/data/2ktic/`这里面就是旧版分笔数据文件
* 在相对URL后面拼接yyyMMdd.zip的格式拼接下载url，比如下载2018年1月31日的分笔数据就是:
`http://www.tdx.com.cn/products/data/data/2ktic/20180131.zip`
