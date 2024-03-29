package roomPkg

import (
	. "gohipernetFake"

	"main/connectedSessions"
	"main/protocol"
)

func (room *baseRoom) _packetProcess_InviteUser(inValidUser *roomUser, packet protocol.Packet) int16 {
	curTime := NetLib_GetCurrnetUnixTime()
	sessionIndex := packet.UserSessionIndex
	sessionUniqueId := packet.UserSessionUniqueId
	NTELIB_LOG_INFO("[[[Room _packetProcess_EnterUser]]]")

	var requestPacket protocol.RoomEnterReqPacket
	(&requestPacket).Decoding(packet.Data)

	userID, ok := connectedSessions.GetUserID(sessionIndex)
	if ok == false {
		_sendRoomEnterResult(sessionIndex, sessionUniqueId, 0, 0, protocol.ERROR_CODE_ENTER_ROOM_INVALID_USER_ID)
		return protocol.ERROR_CODE_ENTER_ROOM_INVALID_USER_ID
	}

	userInfo := addRoomUserInfo{
		userID,
		sessionIndex,
		sessionUniqueId,
	}
	newUser, addResult := room.addUser(userInfo)

	if addResult != protocol.ERROR_CODE_NONE {
		_sendRoomEnterResult(sessionIndex, sessionUniqueId, 0, 0, addResult)
		return addResult
	}

	if connectedSessions.SetRoomNumber(sessionIndex, sessionUniqueId, room.getNumber(), curTime) == false {
		_sendRoomEnterResult(sessionIndex, sessionUniqueId, 0, 0, protocol.ERROR_CODE_ENTER_ROOM_INVALID_SESSION_STATE)
		return protocol.ERROR_CODE_ENTER_ROOM_INVALID_SESSION_STATE
	}

	if room.getCurUserCount() > 1 {
		//룸의 다른 유저에게 통보한다.
		room._sendNewUserInfoPacket(newUser)

		// 지금 들어온 유저에게 이미 채널에 있는 유저들의 정보를 보낸다
		room._sendUserInfoListPacket(newUser)
	}

	roomNumebr := room.getNumber()
	_sendRoomEnterResult(sessionIndex, sessionUniqueId, roomNumebr, newUser.RoomUniqueId, protocol.ERROR_CODE_NONE)
	return protocol.ERROR_CODE_NONE
}

