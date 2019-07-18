package pb

func MakeRsp(mid string, rsp isRsp_Rsp) *Message {
	return &Message{
		Message: &Message_Rsp{
			Rsp: &Rsp{
				Mid: mid,
				Rsp: rsp,
			},
		},
	}
}

func MakeRsp_GetUserInfoRsp(mid string, user *User) *Message {
	return MakeRsp(mid, &Rsp_GetUserInfoRsp{
		GetUserInfoRsp: &GetUserInfoRsp{
			User: user,
		},
	})
}

func MakeRsp_Error(mid, message string) *Message {
	return MakeRsp(mid, &Rsp_Error{
		Error: &Error{
			Message: message,
		},
	})
}

func MakePush(push isPush_Push) *Message {
	return &Message{
		Message: &Message_Push{
			Push: &Push{
				Push: push,
			},
		},
	}
}

func MakePush_ChatPush(message string) *Message {
	return MakePush(&Push_ChatPush{
		ChatPush: &ChatPush{
			Message: message,
		},
	})
}
