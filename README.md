# searchproxy

#加入扫描
使用golang 1.17
扫描全球IP端口，发现
searchproxy目录下执行  make
再 cd docker ，执行 make build_scanport build_proxyscan
编译程序,部署扫描节点 执行  make play_localport 扫描节点，make play_proxyscan 验证端口是否为代理


restful 端口8080，获取可用代理 api 为 GET请求
protocol类型分为 socks5 和 http
google  是否可访问google
/api/get/list?google=true&protocol=socks5&limit=10&skip=0


