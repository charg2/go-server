package main

import (
	. "gohipernetFake"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type EchoServer struct {
	ServerIndex int
	IP          string
	Port        int
}

func createServer(netConfig NetworkConfig) {
	NTELIB_LOG_INFO("CreateServer !!!")

	var server EchoServer

	if server.setIPAddress(netConfig.BindAddress) == false {
		NTELIB_LOG_ERROR("fail. server address")
		return
	}

	// 콜백 함수포인터 등록
	networkFunctor := SessionNetworkFunctors{}
	networkFunctor.OnConnect = server.OnConnect
	networkFunctor.OnReceive = server.OnReceive
	networkFunctor.OnReceiveBufferedData = nil // 무시
	networkFunctor.OnClose = server.OnClose
	networkFunctor.PacketTotalSizeFunc = PacketTotalSize // 패킷의 최대 크기 읽어 들임.
	networkFunctor.PacketHeaderSize = PACKET_HEADER_SIZE
	networkFunctor.IsClientSession = true // serer끼리의 통신인지 cs 통신인지

	// 네트워크 라이브럴 ㅣㅊ ㅗ기화
	NetLibInitNetwork(PACKET_HEADER_SIZE, PACKET_HEADER_SIZE)
	// 서버 네트워크 동작 종료할떄까지 blocking
	NetLibStartNetwork(&netConfig, networkFunctor)

	server.Stop()

}

func (server* EchoServer) setIPAddress(ipAddress string) bool { // string ip port 분리.
	results := strings.Split(ipAddress, ":")
	if len(results) != 2 {
		return false
	}

	server.IP = results[0]
	server.Port, _ = strconv.Atoi(results[1])

	NTELIB_LOG_INFO("Server Addresss", zap.String("IP", server.IP), zap.Int("Port", server.Port))
	return true
}

func (server* EchoServer) OnConnect(sessionIndex int32, sessionUniqueID uint64) {
	NTELIB_LOG_INFO("client OnConnect", zap.Int32("sessionIndex", sessionIndex), zap.Uint64("sessionUniqueId", sessionUniqueID))
}

func (server* EchoServer) OnReceive(sessionIndex int32, sessionUniqueID uint64, data []byte) bool {
	NTELIB_LOG_DEBUG("OnReceive", zap.Int32("sessionIndex", sessionIndex),
		zap.Uint64("sessionUniqueID", sessionUniqueID),
		zap.Int("packetSize", len(data)))

	NetLibISendToClient(sessionIndex, sessionUniqueID, data)
	return true
}

func (server *EchoServer) OnClose(sessionIndex int32, sessionUniqueID uint64) {
	NTELIB_LOG_INFO("client OnCloseClientSession", zap.Int32("sessionIndex", sessionIndex), zap.Uint64("sessionUniqueID", sessionUniqueID))
}

func (server *EchoServer) Stop() {
	NTELIB_LOG_INFO("chatServer Stop !!!")

	NetLib_StopServer() // 이 함수가 꼭 제일 먼저 호출 되어야 한다.

	NTELIB_LOG_INFO("chatServer Stop Waitting...")
	time.Sleep(1 * time.Second)
}
