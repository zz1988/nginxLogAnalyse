package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/astaxie/beego/config"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wangtuanjie/ip17mon"
)

// 初始化IP库
var (
	webRequestMonitor = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "nginx_request",
			Help: "the request summary of nginx",
		},
		[]string{"ipLocation", "requestType", "serviceName", "statusCode"})
)

func init() {
	if err := ip17mon.Init("17monipdb.dat"); err != nil {
		panic(err)
	}
	prometheus.MustRegister(webRequestMonitor)
}

func main() {
	iniconf, _ := config.NewConfig("ini", "nginxAnalyse.conf")
	go analyseNginxLog(iniconf)
	http.Handle("/metrics", promhttp.Handler())
	port := iniconf.String("prometheus::port")
	log.Println(http.ListenAndServe(":"+port, nil))
}

// 分析nginx日志
func analyseNginxLog(iniconf config.Configer) {
	nginxLog := iniconf.String("nginx::log")
	t, err := tail.TailFile(nginxLog, tail.Config{Follow: true})
	if err != nil {
		return
	}
	for line := range t.Lines {
		ipLocation, requestType, requestMethod, statusCode := processLogLine(line.Text)
		// fmt.Println(ip, ipLocation, requestType, requestMethod, statusCode)
		if requestMethod != "" && ipLocation != "" {
			webRequestMonitor.WithLabelValues(ipLocation, requestType, requestMethod, statusCode).Observe(1)
		}
	}
}

//解析nginx日志，返回请求的ip，请求类型，请求方法，返回状态码信息
func processLogLine(tmpLogLine string) (string, string, string, string) {
	ip := strings.TrimSpace(strings.Split(tmpLogLine, "-")[0])
	otherStr := tmpLogLine[strings.Index(tmpLogLine, "]")+1:]
	logInfo := strings.Split(otherStr, "\"")
	requestInfo := strings.Split(logInfo[1], " ")
	requestType := getStringFromArray(requestInfo, 0)
	requestUrl := getStringFromArray(requestInfo, 1)
	requestMethod := requestUrl[strings.LastIndex(requestUrl, "/")+1:]
	if strings.Index(requestUrl, "?") > 0 {
		requestMethod = requestUrl[strings.LastIndex(requestUrl, "/")+1 : strings.Index(requestUrl, "?")]
	}
	statusInfo := strings.Split(strings.TrimSpace(logInfo[2]), " ")
	statusCode := getStringFromArray(statusInfo, 0)
	ipLocation := getLocationByIP(ip)
	return ipLocation, requestType, requestMethod, statusCode
}

// 根据ip获取地址
func getLocationByIP(ip string) string {
	ipLocation, err := ip17mon.Find(ip)
	if err != nil {
		fmt.Println("err:", err)
		return ""
	}
	return getEnNameOfProvince(ipLocation.Region)
}

// 从字符串数组中获取对应index的字符串，如果找不到，返回空字符串
func getStringFromArray(strArray []string, index int) string {
	resultStr := ""
	if len(strArray) > index {
		resultStr = strArray[index]
	}
	return resultStr
}

// 获取省份名称的英文名
func getEnNameOfProvince(srcProvinceName string) string {
	switch srcProvinceName {
	case "甘肃":
		return "gansu"
	case "青海":
		return "qinghai"
	case "四川":
		return "sichuan"
	case "河北":
		return "hebei"
	case "云南":
		return "yunnan"
	case "贵州":
		return "guizhou"
	case "湖北":
		return "hubei"
	case "河南":
		return "henan"
	case "山东":
		return "shandong"
	case "江苏":
		return "jiangsu"
	case "安徽":
		return "anhui"
	case "浙江":
		return "zhejiang"
	case "江西":
		return "jiangxi"
	case "福建":
		return "fujian"
	case "广东":
		return "guangdong"
	case "湖南":
		return "hunan"
	case "海南":
		return "hainan"
	case "辽宁":
		return "liaoning"
	case "吉林":
		return "jilin"
	case "黑龙江":
		return "heilongjiang"
	case "山西":
		return "shanxi"
	case "陕西":
		return "shaanxi"
	case "台湾":
		return "taiwan"
	case "北京":
		return "beijing"
	case "上海":
		return "shanghai"
	case "重庆":
		return "chongqing"
	case "天津":
		return "tianjing"
	case "内蒙古":
		return "neimenggu"
	case "广西":
		return "guangxi"
	case "西藏":
		return "xizang"
	case "宁夏":
		return "ningxia"
	case "新疆":
		return "xinjiang"
	case "香港":
		return "xianggang"
	case "澳门":
		return "aomen"
	}
	return ""
}
