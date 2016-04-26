# GoConf
配置格式转换工具，支持xlxs,csv,lua,json等的相互转换
## 功能说明
* xlsx与csv之间的相互转换，必须指定xlsx的sheetname
* json与lua之间的相互转换，lua的table不能混合保护数组和键值对
* xlsx/csv转换成json/lua
 - 指定一列为key，转换后每行为一个键值对，键为key对应列的值，值为table（其中列名为key）
 - 不指定key，转换后每行对应数组的一个元素，值为table
* json/lua转换成xlsx/csv，json/lua的格式必须与上述情况匹配

## 使用方式
./GoConf -i input_dir -o output_dir -it [xlsx|csv|lua|json] -ot [xlsx|csv|lua|json] -k column - sheet
