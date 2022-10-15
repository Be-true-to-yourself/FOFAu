## FOFAu代码解析

#### 0x01 设计思路

​	首先，写一个FOFA客户端，想到的第一件事就是需要调用哪些库，首当其冲，肯定是net库中的http库函数；其次，通过观察FOFA官网，我们应该知道有一个base64加密,所以，第二个需要用到的库是encoding库中的base64库函数；第三个，必备库函数，输入输出——fmt库和log库；第四个，要想做到命令行的参数输入效果，所使用的库——flag库。

​	到此，一个基本的设计思路已经确定，收到参数->http请求FOFA API->接受数据->处理数据->输出

#### 0x02 代码解析

​		首先我们需要做一个存储在FOFA注册过的email和FOFA KEY，那么我们需要创建一个文件。这里提一下fmt.Scanln和fmt.Scanf的区别——fmt.Scanln()是一行行的输入用enter结束,fmt.Scanf()是所有的输入,中间用空格隔开,enter结束。所以这里小CV使用的是fmt.Scanln()函数，而不是fmt.Scanf()函数。这里还使用了os库，具体使用可以参考Go中文API库：https://studygolang.com/pkgdoc，有兴趣的师傅可以去查看。

```go
//创建文件
func CreatFile(email string, key string) {
	log.Println("存放FOFA KEY的文本文档fofa_key.txt不存在")
	fmt.Print("请输入您在FOFA注册的email地址:")
	fmt.Scanln(&email)
	fmt.Print("请输入您在FOFA获取的API KEY:")
	fmt.Scanln(&key)
	fofa_key_data := fmt.Sprintf("FOFA_Email=%s\nFOFA_Key=%s", email, key)
	file_creat, err := os.Create(fofa_file_key)
	if err != nil {
		log.Fatalln("创建文件失败，程序终止")
	}
	file_creat.WriteString(fofa_key_data)
	log.Println("文件已创建并且FOFA KEY已写入")
	file_creat.Close()
}
```

​		这里解释一下这个fofa_file_key的由来——提前设置好的静态变量。

```go
const fofa_file_key = "./fofa_key.txt"
```

​		创建文件之后肯定需要读取文件，所以我们需要写一个读取文件的函数。这里又用到了一个库——regexp库，主要作用用于正则表达式搜索。感兴趣是师傅可以点击上面推荐的链接查看函数库的使用。关于读取文件的写法，师傅们可以参考该条链接：https://blog.csdn.net/slphahaha/article/details/122995124。正则表达式，第一个`FOFA_Email=\w+@\w+.com`，其中\w+@\w+.com正则表达式的意思是获取email；第二个，`FOFA_Key=\w{0,32}`，其中\w{0,32}正则表达式的意思是获取字母、数字、下划线，获取数量为0到32个（因为FOFA KEY是32位）。最后，if判断获取内容，若为不为空，则返回email和key的内容；若为空，则if不执行，执行最后语句，返回nil字符串。

```go
//读取FOFA KEY
func ReadFile() (email string, key string) {
	file, err := os.Open(fofa_file_key)
	if err != nil {
		log.Fatalln("文件打开失败")
	}
    //创建一个大小为1024个字节变量数组
	read_buf := make([]byte, 1024)
	for {
        //每次读取1024个字节
		len, _ := file.Read(read_buf)
		if len == 0 {
			break
		}
	}
	file.Close()
	fofa_key_data := string(read_buf)
	//使用正则获取关键段落
	email_match := regexp.MustCompile(`FOFA_Email=\w+@\w+.com`)
	key_match := regexp.MustCompile(`FOFA_Key=\w{0,32}`)
	email_value := email_match.FindStringSubmatch(fofa_key_data)
	key_value := key_match.FindStringSubmatch(fofa_key_data)
	if email_value != nil && key_value != nil {
	//去除获取的标头
		email = strings.Replace(email_value[0], "FOFA_Email=", "", -1)
		key = strings.Replace(key_value[0], "FOFA_Key=", "", -1)
		return email, key
	}
	return "nil", "nil"
}
```

​		获取FOFA邮箱和FOFA KEY的功能函数已经完成，那么我们需要做的就是对FOFA API进行访问，获取数据。这里又使用了一个库——ioutil库，这个库的功能函数ioutil.ReadAll()作用是从res.Body读取数据直到EOF或遇到error，返回读取的数据和遇到的错误，成功的调用返回的err为nil而非EOF。

```go
//向FOFA发出请求
func get_fofa_contents(url string) (result string) {
	log.Print("正在访问链接", url)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal("请求出错，请查看填写的email和FOFA Key,错误参数为:", err)
	}
	fofa_data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("收到的FOFA数据类型错误")
	}
	return string(fofa_data)
}
```

​		有写过FOFA客户端的师傅应该知道，从FOFA的API中获取的数据是JSON格式的数据，所以我们下一步应该是写一个JSON格式转换的功能函数。

​		这里提一下FOFA API返回的JSON格式

```json
{
    "error": false, // 是否出现错误
    "size": 查询总数量,
    "page": 当前页码, 
    "mode": "extended",
    "query": "查询语句", 
    "results": [
      [
      数据1
      ],
      [
      数据2
      ],
      [
      ......
      ]
    ]
  }
```

​		在此功能函数中用到了json库，json.Unmarshl()库的作用是将JSON字符串转换成定义类型的切片，比如，小CV这里创建了一个map类型，用于接受JSON字符串。值得一提的是动态创建二维数组的方法，妙用make函数`var arr = make([][]string, len(results_1))`，刚学Go的师傅们可以借鉴。有的师傅肯定疑惑，这里为什么要创建二维数组，不是已经将results字段里面的内容提取出来了吗？这里小CV解释一下，提取了results字段，现在里面还是一个整体的字符串，而我们需要一个二维数组将里面的数据分离出来，所以小CV这里创建了一个二维数组，而为什么采用动态创建二维数组，主要是考虑师傅们使用的时候，获取的参数数据自由度，比如，默认是获取ip,host,port这三个字段，那么二维数组长度就是一级长度就是3。也就是这里为什么是`len(results_1)`的缘故。

​		在此，JSON数据转换就完成了，小CV也是初学Go语言，所以对JSON数据的处理比较笨，大师傅们轻点喷。

```go
//转换JSON数据
func Convert_json(json_result string) (results [][]string) {
	data_convert := make(map[string]interface{})
	if err := json.Unmarshal([]byte(json_result), &data_convert); err != nil {
		log.Fatal("JSON数据解析失败")
	}
	//动态创建二维数组
	results_1 := data_convert["results"].([]interface{})
	var arr = make([][]string, len(results_1))
	i := 0
	for _, result_1 := range results_1 {
		result_2 := result_1.([]interface{})
		arr[i] = make([]string, len(result_2))
		j := 0
		for _, result_3 := range result_2 {
			arr[i][j] = fmt.Sprintf("%v", result_3)
			j++
		}
		i++
	}
	return arr
}
```

​		有了结果数据，那么我们肯定就是输出结果数据，所以下面就是写一个输出功能函数。

```go
//输出查询数据
func out_result(results [][]string, columns []string, dict map[string]string) {
	log.Println("FOFA查询结果如下：")
	for i := 0; i < len(results); i++ {
		for j := 0; j < len(columns); j++ {
			fmt.Printf("%s:%s\n", dict[columns[j]], results[i][j])
		}
	}
}
```

​		这里使用了一个map类型的变量，主要作用是翻译字段（FOFA官网的API文档里面有写）。

```go
dict := map[string]string{"ip": "ip地址", "port": "端口", "protocol": "协议名", "country": "国家代码", "country_name": "国家名", "region": "区域", "city": "城市", "longitude": "地理位置-经度", "latitude": "地理位置-纬度", "as_number": "asn编号", "as_organization": "asn组织", "host": "主机名", "domain": "域名", "os": "操作系统", "server": "网站server", "icp": "icp备案号", "title": "网站标题", "jarm": "jarm 指纹", "header": "网站header", "banner": "协议 banner", "cert": "证书", "body": "网站正文内容", "fid": "fid", "structinfo": "结构化信息"}
```

​		写到这里基本上FOFA客户端的功能已经实现了，当然，小CV这里还写了一个保存数据生成csv文件的功能函数，让使用的师傅体验感更好。所以，用到了encoding库中的csv功能函数。师傅们可以参考https://blog.csdn.net/weixin_44676081/article/details/108029175，用于查看csv功能函数的使用方法。

```go
//保存结果
func ResultSave(results [][]string, columns []string, dict map[string]string, save_path string) {
	file, err := os.Create(save_path)
	if err != nil {
		log.Fatal("创建保存数据文件失败，失败原因：", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
    //写字段名
	var column_name = make([]string, len(columns))
	for i, column := range columns {
		column_name[i] = dict[column]
	}
	writer.Write(column_name)
    //写内容数据
	var column_value = make([]string, len(results[0]))
	for _, result_1 := range results {
		for j, result_2 := range result_1 {
			column_value[j] = result_2
		}
		writer.Write(column_value)
	}
	writer.Flush()
	log.Printf("文件已保存，保存路径：%s", save_path)
}
```

​		最后使用main主函数进行调用。提一下log的使用，log.Fatal()函数是输出之后结束整个进程的使用，log.SetPrefix()函数设置输出头，比如，[CV].......这种格式，小CV这里主要是用于署名，hhhhh

```go
func main() {
	var email string
	var key string
	fmt.Println(flg)
	log.SetPrefix("[CV]")
	//判断文件是否存在
	_, err := os.Stat(fofa_file_key)
	if os.IsNotExist(err) {
		CreatFile(email, key)
		log.Fatal("使用本工具请输入参数-h 查看帮助信息")
	} else {
		for {
			email, key = ReadFile()
			if email == "nil" || key == "nil" {
				log.Println("未获取到Email或FOFA KEY")
				os.Remove(fofa_file_key)
				CreatFile(email, key)
			} else {
				break
			}
		}
	}
	//设置工具参数
	content_point := flag.String("q", ``, "参数q：搜索内容，默认为空（必填项）")
	size := flag.String("s", "100", "参数s：获取搜索的数量，默认为100条，最大为10000条")
	field := flag.String("f", "ip,host,port", "	参数f：获取指定队列名，默认ip,host,port")
	save := flag.String("sv", "up", "参数sv：搜索结果保存，默认开启")
	save_path := flag.String("p", "_"+time.Now().Format("2006_01_02_15_04_05"), "参数p：结果保存位置以及文件名，默认本文件夹，如：./xxxx")
	flag.Parse()
	//判断q参数是否为空
	if *content_point == `` {
		log.Fatal("搜索参数为空，请填写q参数")
	}
	//生成url
	dict := map[string]string{"ip": "ip地址", "port": "端口", "protocol": "协议名", "country": "国家代码", "country_name": "国家名", "region": "区域", "city": "城市", "longitude": "地理位置-经度", "latitude": "地理位置-纬度", "as_number": "asn编号", "as_organization": "asn组织", "host": "主机名", "domain": "域名", "os": "操作系统", "server": "网站server", "icp": "icp备案号", "title": "网站标题", "jarm": "jarm 指纹", "header": "网站header", "banner": "协议 banner", "cert": "证书", "body": "网站正文内容", "fid": "fid", "structinfo": "结构化信息"}
	*save_path = "./" + *content_point + *save_path + ".csv"  //生成文件名拼接
	columns := strings.Split(*field, ",")
	content := base64.StdEncoding.EncodeToString([]byte(*content_point))
	content = strings.Replace(content, "==", "", -1)
	url := fmt.Sprintf("%semail=%s&key=%s&qbase64=%s&size=%s&field=%s", fofaurl, email, key, content, *size, *field)
	//开始对FOFA的数据进行收集
	result_json := get_fofa_contents(url)
	log.Println("搜索内容为：", *content_point)
	results := Convert_json(result_json)
	out_result(results, columns, dict)
	if *save == "up" {
		ResultSave(results, columns, dict, *save_path)
	}
	if *save != "off" && *save != "up" {
		log.Fatal("sv参数错误，只能选择up或者off")
	}

}
```

#### 0x03 申明

​		**此工具仅供学习网络安全交流，不负有相关法律责任。请大家维护好网络环境，请勿进行非法测试。**

​		**以上内容属于原创，转发请附上链接。**



​		工具已经开源，可以关注公众号：乌托邦安全团队，在后台回复：FOFAu，可获取GitHub链接。