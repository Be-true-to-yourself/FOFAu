package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const fofaurl = "https://fofa.info/api/v1/search/all?"

const fofa_file_key = "./fofa_key.txt"

const flg = `
___________________  ________________         
\_   _____/\_____  \ \_   _____/  _  \  __ __ 
 |    __)   /   |   \ |    __)/  /_\  \|  |  \
 |     \   /    |    \|     \/    |    \  |  /
 \___  /   \_______  /\___  /\____|__  /____/ 
     \/            \/     \/         \/        v1.0
                                         --作者 乌托邦安全团队 CV
`

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

//读取FOFA KEY
func ReadFile() (email string, key string) {
	file, err := os.Open(fofa_file_key)
	if err != nil {
		log.Fatalln("文件打开失败")
	}
	read_buf := make([]byte, 1024)
	for {
		len, _ := file.Read(read_buf)
		if len == 0 {
			break
		}
	}
	file.Close()
	fofa_key_data := string(read_buf)
	email_match := regexp.MustCompile(`FOFA_Email=\w+@\w+.com`)
	key_match := regexp.MustCompile(`FOFA_Key=\w{0,32}`)
	email_value := email_match.FindStringSubmatch(fofa_key_data)
	key_value := key_match.FindStringSubmatch(fofa_key_data)
	if email_value != nil && key_value != nil {
		email = strings.Replace(email_value[0], "FOFA_Email=", "", -1)
		key = strings.Replace(key_value[0], "FOFA_Key=", "", -1)
		return email, key
	}
	return "nil", "nil"
}

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

//输出查询数据
func out_result(results [][]string, columns []string, dict map[string]string) {
	log.Println("FOFA查询结果如下：")
	for i := 0; i < len(results); i++ {
		for j := 0; j < len(columns); j++ {
			fmt.Printf("%s:%s\n", dict[columns[j]], results[i][j])
		}
	}
}

//保存结果
func ResultSave(results [][]string, columns []string, dict map[string]string, save_path string) {
	file, err := os.Create(save_path)
	if err != nil {
		log.Fatal("创建保存数据文件失败，失败原因：", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	var column_name = make([]string, len(columns))
	for i, column := range columns {
		column_name[i] = dict[column]
	}
	writer.Write(column_name)
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
	*save_path = "./" + *content_point + *save_path + ".csv"
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
