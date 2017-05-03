package main

import (
	"log"
	"net/http"

	"github.com/astaxie/beego/config"
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wangtuanjie/ip17mon"
	"regexp"
	"strconv"
	"strings"
)

const (
	//正则匹配，根据正则表达式抽取有用信息
	regexPhase = `([^"]*) - - \[.*\] "([^"]*) /CreditFunc/v2.1/([^"]*) .*" "www.miniscores.cn:([^"]*)" .* "([^"]*)" ".*" ".*" ".*" ".* "([^"]*)" "([^"]*)" "([^"]*)"`
)

// 初始化IP库
var (
	// 记录每个查询请求的耗时
	webRequestGuage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nginx_forwarding_guage",
			Help: "the request costTime of nginx",
		},
		[]string{"ipLocation", "requestType", "requestPort", "serviceName", "statusCode", "remoteServer"})

	// 记录每个查询请求的数目
	webRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nginx_forwarding",
			Help: "the request count of nginx",
		},
		[]string{"ipLocation", "requestType", "requestPort", "serviceName", "statusCode", "remoteServer"})
)

func init() {
	if err := ip17mon.Init("17monipdb.dat"); err != nil {
		panic(err)
	}
	prometheus.MustRegister(webRequestGuage)
	prometheus.MustRegister(webRequestCounter)
}

func main() {
	iniconf, _ := config.NewConfig("ini", "nginxAnalyse.conf")
	// 分析nginx日志
	go func() {
		nginxLog := iniconf.String("nginx::log")
		t, err := tail.TailFile(nginxLog, tail.Config{Follow: true, ReOpen: true, Poll: true})
		if err != nil {
			return
		}
		for line := range t.Lines {
			ipLocation, requestType, requestPort, serviceName, statusCode, remoteServer, costTimeAll, _, ok := processLogLine(line.Text)
			if ok {
				webRequestGuage.WithLabelValues(ipLocation, requestType, requestPort, serviceName, statusCode, remoteServer).Set(costTimeAll)
				webRequestCounter.WithLabelValues(ipLocation, requestType, requestPort, serviceName, statusCode, remoteServer).Inc()
			}
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	port := iniconf.String("prometheus::port")
	log.Println(http.ListenAndServe(":"+port, nil))
}

//解析nginx日志，返回请求的ip，请求类型，请求端口号，请求方法，返回状态码信息，请求目标地址，请求总耗时，请求单纯耗时
func processLogLine(tmpLogLine string) (string, string, string, string, string, string, float64, float64, bool) {
	re := regexp.MustCompile(regexPhase)
	segs2 := re.FindAllStringSubmatch(tmpLogLine, -1)
	if len(segs2) > 0 && len(segs2[0]) == 9 {
		srcIP := segs2[0][1]                                     //请求地址ip
		requestType := segs2[0][2]                               //请求方法
		serviceName := segs2[0][3]                               //请求服务名
		requestPort := segs2[0][4]                               //请求端口
		statusCode := segs2[0][5]                                //请求状态码
		remoteServer := segs2[0][6]                              //转发服务地址
		costTimeAll, _ := strconv.ParseFloat(segs2[0][7], 3)     //总请求耗时
		costTimeRequest, _ := strconv.ParseFloat(segs2[0][8], 3) //单独请求耗时
		ipLocation := getLocationByIP(srcIP)
		if strings.Contains(serviceName, "?") { //如果服务名中包含？，则认为无效，不统计
			return ipLocation, requestType, requestPort, serviceName, statusCode, remoteServer, costTimeAll, costTimeRequest, false
		}
		return ipLocation, requestType, requestPort, serviceName, statusCode, remoteServer, costTimeAll, costTimeRequest, true
	}
	return "", "", "", "", "", "", 0, 0, false
}

// 根据ip获取地址
func getLocationByIP(ip string) string {
	ipLocation, err := ip17mon.Find(ip)
	if err != nil {
		log.Println("err:", err)
		return ""
	}
	return getEnNameOfProvince(ipLocation.Region)
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
