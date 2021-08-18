package action

var (
	KEY_ARRAY         = "[]"
	KEY_COUNT         = "count"
	KEY_PAGE          = "page"
	KEY_JOIN          = "join"
	KEY_CONTION       = "{}"
	KEY_LIST          = "<>"
	KEY_FOREIGN       = "@"
	KEY_REG           = "~"
	KEY_LIKE          = "$"
	KEY_ADD           = "+"
	KEY_REDUCE        = "-"
	KEY_RANGE         = "%" //"date%":"2017-10-01,2018-10-01"
	KEY_GREATER       = ">"
	KEY_LESS          = "<"
	KEY_GREATER_EQUAL = ">="
	KEY_LESS_EQUAL    = "<="
	KEY_SET           = "@set"
	KEY_COLUMN        = "@column"
	KEY_ORDER         = "@order"
	KEY_COMBINE       = "@combine"
	KEY_TOTAL         = "total@"
	KEY_OK            = "ok"
	KEY_CODE          = "code"
	KEY_MSG           = "msg"

	//
	CODE_SUCCESS               = 200 //成功
	CODE_UNSUPPORTED_ENCODING  = 400 //编码错误
	CODE_ILLEGAL_ACCESS        = 401 //权限错误
	CODE_UNSUPPORTED_OPERATION = 403 //禁止操作
	CODE_NOT_FOUND             = 404 //未找到
	CODE_ILLEGAL_ARGUMENT      = 406 //参数错误
	CODE_NOT_LOGGED_IN         = 407 //未登录
	CODE_TIME_OUT              = 408 //超时
	CODE_CONFLICT              = 409 //重复，已存在
	CODE_CONDITION_ERROR       = 412 //条件错误，如密码错误
	CODE_UNSUPPORTED_TYPE      = 415 //类型错误
	CODE_OUT_OF_RANGE          = 416 //超出范围
	CODE_NULL_POINTER          = 417 //对象为空
	CODE_SERVER_ERROR          = 500 //服务器内部错误

	//
	MSG_SUCCEED      = "success"                //成功
	MSG_SERVER_ERROR = "Internal Server Error!" //服务器内部错误
)
