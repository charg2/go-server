package main

import (
	"flag" // cmd 기본인자 파싱 라이브러리.
	. "gohipernetFake"
)

func main() {
	NetLibInitLog() // 네트워크 라이브러리 초기화

	netConfigClient := parseAppConfig() // cmd 인자 파싱해서 서버 설정정보
	netConfigClient.WriteNetworkConfig(true)

	createServer(netConfigClient)
}

func parseAppConfig() NetworkConfig {
	NTELIB_LOG_INFO("[[Setting NetworkConfig]]")

	client := NetworkConfig{}

	flag.BoolVar(&client.IsTcp4Addr, "c_IsTcp4Addr", true, "bool flag")
	flag.StringVar(&client.BindAddress, "c_BindAddress", "127.0.0.1:11021", "string flag")
	flag.IntVar(&client.MaxSessionCount, "c_MaxSessionCount", 0, "int flag")
	flag.IntVar(&client.MaxPacketSize, "c_MaxPacketSize", 0, "int flag")

	flag.Parse()
	return client
}
