# nginxLogAnalyse
Use Prometheus analyse the nginx log

# nginx日志分析
此程序读取nginx日志信息，将其汇总统计到prometheus的解析端口；
分为以下几步：
1. 读取nginx的access.log日志，获取当前请求的ip，requestType(POST/GET),请求的Url(解析出服务名),statusCode(请求状态);
2. 解析ip，通过一个根据ip查询地址信息的库(https://www.ipip.net/)，查得当前请求ip所属的省份信息；
3. 将这些信息作为label，统一汇总到prometheus中统计；
4. 配置文件nginxAnalyse.conf只设置prometheus的统计端口号和日志位置

# Garafana配置
Garafana需要安装以下插件：（命令行输入）
grafana-cli plugins install grafana-piechart-panel
grafana-cli plugins install grafana-worldmap-panel

使用地图插件时，需要设置地图中的标记点；修改目录下的states.json文件，替换为当前目录下的states.json
$PROMETHEUS/data/plugins/grafana-worldmap-panel/src/data
$PROMETHEUS/data/plugins/grafana-worldmap-panel/dist/data
Note:原始的地图插件所包含的标记点是有限的，需要我们设置中国的地图坐标；
