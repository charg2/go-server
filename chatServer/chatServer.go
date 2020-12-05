package main

import (
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	. "gohipernetFake"

	"main/connectedSessions"
	"main/protocol"
	"main/roomPkg"
)


type configAppServer struct {
	GameName                   string

	RoomMaxCount               int32
	RoomStartNum               int32
	RoomMaxUserCount           int32
}

type ChatServer struct {
	ServerIndex int
	IP          string
	Port        int

	PacketChan	chan protocol.Packet //

	RoomMgr *roomPkg.RoomManager
}

// 채팅 서버 시작 지점.
func createAnsStartServer(netConfig NetworkConfig, appConfig configAppServer) {
	NTELIB_LOG_INFO("CreateServer !!!")

	var server ChatServer

	if server.setIPAddress(netConfig.BindAddress) == false {
		NTELIB_LOG_ERROR("fail. server address")
		return
	}

	protocol.Init_packet() // 패킷 헤더 크기 초기화

	maxUserCount := appConfig.RoomMaxCount * appConfig.RoomMaxUserCount // 최대 동접
	connectedSessions.Init(netConfig.MaxSessionCount, maxUserCount) // 세션 매니저 초기화

	server.PacketChan = make(chan protocol.Packet, 256) // 채널 크기 256, 이걸 안정하면 blocking 된다.

	roomConfig := roomPkg.RoomConfig{
		appConfig.RoomStartNum,
		appConfig.RoomMaxCount,
		appConfig.RoomMaxUserCount,
	}
	server.RoomMgr = roomPkg.NewRoomManager(roomConfig)


	// 패킷 처리 고루틴 함수 실행
	go server.PacketProcess_goroutine()


	networkFunctor := SessionNetworkFunctors{}
	networkFunctor.OnConnect = server.OnConnect
	networkFunctor.OnReceive = server.OnReceive
	networkFunctor.OnReceiveBufferedData = nil
	networkFunctor.OnClose = server.OnClose
	networkFunctor.PacketTotalSizeFunc = PacketTotalSize
	networkFunctor.PacketHeaderSize = PACKET_HEADER_SIZE
	networkFunctor.IsClientSession = true


	NetLibInitNetwork(PACKET_HEADER_SIZE, PACKET_HEADER_SIZE)
	NetLibStartNetwork(&netConfig, networkFunctor) // 서버가 시작 되는 부분 accept 걸고, concurrent 모델 처럼 고루틴을 생성해서 IO를 시작한다.

	server.Stop()
}

func (server *ChatServer) setIPAddress(ipAddress string) bool {
	results := strings.Split(ipAddress, ":")
	if len(results) != 2 {
		return false
	}

	server.IP = results[0]
	server.Port, _ = strconv.Atoi(results[1])

	NTELIB_LOG_INFO("Server Address", zap.String("IP", server.IP), zap.Int("Port", server.Port))
	return true
}

func (server *ChatServer) Stop() {
	NTELIB_LOG_INFO("chatServer Stop !!!")

	NetLib_StopServer() // 이 함수가 꼭 제일 먼저 호출 되어야 한다.

	NTELIB_LOG_INFO("chatServer Stop Waitting...")
	time.Sleep(1 * time.Second)
}

// 연결 됬을시 세선 매니저에 추가.
func (server *ChatServer) OnConnect(sessionIndex int32, sessionUniqueID uint64) {
	NTELIB_LOG_INFO("client OnConnect", zap.Int32("sessionIndex",sessionIndex), zap.Uint64("sessionUniqueId",sessionUniqueID))

	connectedSessions.AddSession(sessionIndex, sessionUniqueID)
}

// 데이터 수신시 로그를 직고
// 바이트 배열의  패킬을 만들어 채널로 전송
func (server *ChatServer) OnReceive(sessionIndex int32, sessionUniqueID uint64, data []byte) bool {
	NTELIB_LOG_DEBUG("OnReceive", zap.Int32("sessionIndex", sessionIndex),
		zap.Uint64("sessionUniqueID", sessionUniqueID),
		zap.Int("packetSize", len(data)))

	server.DistributePacket(sessionIndex, sessionUniqueID, data)
	return true
}


func (server *ChatServer) OnClose(sessionIndex int32, sessionUniqueID uint64) {
	NTELIB_LOG_INFO("client OnCloseClientSession", zap.Int32("sessionIndex", sessionIndex), zap.Uint64("sessionUniqueId", sessionUniqueID))

	server.disConnectClient(sessionIndex, sessionUniqueID)
}

// 세션 상태에 따라 종료 처리를 한다.
func (server *ChatServer) disConnectClient(sessionIndex int32, sessionUniqueId uint64) {
	// 로그인도 안한 유저라면 그냥 여기서 처리한다.
	// 방 입장을 안한 유저라면 여기서 처리해도 괜찮지만 아래로 넘긴다.
	if connectedSessions.IsLoginUser(sessionIndex) == false {
		NTELIB_LOG_INFO("DisConnectClient - Not Login User", zap.Int32("sessionIndex", sessionIndex))
		connectedSessions.RemoveSession(sessionIndex, false)
		return
	}


	packet := protocol.Packet {
		sessionIndex,
		sessionUniqueId,
		protocol.PACKET_ID_SESSION_CLOSE_SYS,
		0,
		nil,
	}

	server.PacketChan <- packet

	NTELIB_LOG_INFO("DisConnectClient - Login User", zap.Int32("sessionIndex", sessionIndex))
}
