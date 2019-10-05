## 数据导入与公式转换工具使用方法

### 自动上传、分析、插入

现在网页前端UI已经支持编辑TableMetadata、DatabaseTable、导入xlsx/csv数据，请直接前往使用。

新版逻辑在data_file_parser文件夹下实现，以下内容为旧版说明。

### (Deprecated) xlsx数据导入

编译2d_table_creater并运行。

本程序能够自动判断数据是一维表还是二维表并导入。

参数：

-e="二维表列名称"，目前出现过的有EWBHXH和EWBLXH

-f="单个xlsx文件路径"，若指定，则仅导入该文件，忽略-d参数

-d="xlsx目录路径"，该目录只能包含xlsx文件，全部导入

### (Deprecated) csv数据导入

编译2d_table_creater_csv并运行。

用法与xlsx数据导入程序相同。

### (Deprecated) 所得税数据导入（带表名、列名映射）

编译sds_table_creater并运行。

同目录下放sds_mapping.json，提供映射信息（该json可以由目录中提供的两个py程序parse he14,he15的mapping得到）。

用法与xlsx数据导入程序相同。

### 公式格式更新

编译pipeline_converter并运行。

从 年月日分离 更新到 Timstamp存储，Year/Month/Day运算提取。

参数：

-start 起始下标（设为0更新全部公式）

-limit 更新数量（设为一个较大值更新全部公式）
