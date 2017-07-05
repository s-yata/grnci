package grnci

import (
	"encoding/json"
	"net/http"
)

// Error codes.
const (
	InvalidAddress = 1000 + iota
	InvalidCommand
	InvalidOperation
	InvalidResponse
	InvalidType
	NetworkError
	UnknownError
)

// getCodeText returns a string that briefly describes the specified code.
// getCodeText supports Groonga result codes (C.grn_rc) [,0],
// Grnci error codes [1000,] and HTTP status codes [100,999].
func getCodeText(code int) string {
	switch code {
	case 0:
		return "GRN_SUCCESS"
	case 1:
		return "GRN_END_OF_DATA"
	case -1:
		return "GRN_UNKNOWN_ERROR"
	case -2:
		return "GRN_OPERATION_NOT_PERMITTED"
	case -3:
		return "GRN_NO_SUCH_FILE_OR_DIRECTORY"
	case -4:
		return "GRN_NO_SUCH_PROCESS"
	case -5:
		return "GRN_INTERRUPTED_FUNCTION_CALL"
	case -6:
		return "GRN_INPUT_OUTPUT_ERROR"
	case -7:
		return "GRN_NO_SUCH_DEVICE_OR_ADDRESS"
	case -8:
		return "GRN_ARG_LIST_TOO_LONG"
	case -9:
		return "GRN_EXEC_FORMAT_ERROR"
	case -10:
		return "GRN_BAD_FILE_DESCRIPTOR"
	case -11:
		return "GRN_NO_CHILD_PROCESSES"
	case -12:
		return "GRN_RESOURCE_TEMPORARILY_UNAVAILABLE"
	case -13:
		return "GRN_NOT_ENOUGH_SPACE"
	case -14:
		return "GRN_PERMISSION_DENIED"
	case -15:
		return "GRN_BAD_ADDRESS"
	case -16:
		return "GRN_RESOURCE_BUSY"
	case -17:
		return "GRN_FILE_EXISTS"
	case -18:
		return "GRN_IMPROPER_LINK"
	case -19:
		return "GRN_NO_SUCH_DEVICE"
	case -20:
		return "GRN_NOT_A_DIRECTORY"
	case -21:
		return "GRN_IS_A_DIRECTORY"
	case -22:
		return "GRN_INVALID_ARGUMENT"
	case -23:
		return "GRN_TOO_MANY_OPEN_FILES_IN_SYSTEM"
	case -24:
		return "GRN_TOO_MANY_OPEN_FILES"
	case -25:
		return "GRN_INAPPROPRIATE_I_O_CONTROL_OPERATION"
	case -26:
		return "GRN_FILE_TOO_LARGE"
	case -27:
		return "GRN_NO_SPACE_LEFT_ON_DEVICE"
	case -28:
		return "GRN_INVALID_SEEK"
	case -29:
		return "GRN_READ_ONLY_FILE_SYSTEM"
	case -30:
		return "GRN_TOO_MANY_LINKS"
	case -31:
		return "GRN_BROKEN_PIPE"
	case -32:
		return "GRN_DOMAIN_ERROR"
	case -33:
		return "GRN_RESULT_TOO_LARGE"
	case -34:
		return "GRN_RESOURCE_DEADLOCK_AVOIDED"
	case -35:
		return "GRN_NO_MEMORY_AVAILABLE"
	case -36:
		return "GRN_FILENAME_TOO_LONG"
	case -37:
		return "GRN_NO_LOCKS_AVAILABLE"
	case -38:
		return "GRN_FUNCTION_NOT_IMPLEMENTED"
	case -39:
		return "GRN_DIRECTORY_NOT_EMPTY"
	case -40:
		return "GRN_ILLEGAL_BYTE_SEQUENCE"
	case -41:
		return "GRN_SOCKET_NOT_INITIALIZED"
	case -42:
		return "GRN_OPERATION_WOULD_BLOCK"
	case -43:
		return "GRN_ADDRESS_IS_NOT_AVAILABLE"
	case -44:
		return "GRN_NETWORK_IS_DOWN"
	case -45:
		return "GRN_NO_BUFFER"
	case -46:
		return "GRN_SOCKET_IS_ALREADY_CONNECTED"
	case -47:
		return "GRN_SOCKET_IS_NOT_CONNECTED"
	case -48:
		return "GRN_SOCKET_IS_ALREADY_SHUTDOWNED"
	case -49:
		return "GRN_OPERATION_TIMEOUT"
	case -50:
		return "GRN_CONNECTION_REFUSED"
	case -51:
		return "GRN_RANGE_ERROR"
	case -52:
		return "GRN_TOKENIZER_ERROR"
	case -53:
		return "GRN_FILE_CORRUPT"
	case -54:
		return "GRN_INVALID_FORMAT"
	case -55:
		return "GRN_OBJECT_CORRUPT"
	case -56:
		return "GRN_TOO_MANY_SYMBOLIC_LINKS"
	case -57:
		return "GRN_NOT_SOCKET"
	case -58:
		return "GRN_OPERATION_NOT_SUPPORTED"
	case -59:
		return "GRN_ADDRESS_IS_IN_USE"
	case -60:
		return "GRN_ZLIB_ERROR"
	case -61:
		return "GRN_LZ4_ERROR"
	case -62:
		return "GRN_STACK_OVER_FLOW"
	case -63:
		return "GRN_SYNTAX_ERROR"
	case -64:
		return "GRN_RETRY_MAX"
	case -65:
		return "GRN_INCOMPATIBLE_FILE_FORMAT"
	case -66:
		return "GRN_UPDATE_NOT_ALLOWED"
	case -67:
		return "GRN_TOO_SMALL_OFFSET"
	case -68:
		return "GRN_TOO_LARGE_OFFSET"
	case -69:
		return "GRN_TOO_SMALL_LIMIT"
	case -70:
		return "GRN_CAS_ERROR"
	case -71:
		return "GRN_UNSUPPORTED_COMMAND_VERSION"
	case -72:
		return "GRN_NORMALIZER_ERROR"
	case -73:
		return "GRN_TOKEN_FILTER_ERROR"
	case -74:
		return "GRN_COMMAND_ERROR"
	case -75:
		return "GRN_PLUGIN_ERROR"
	case -76:
		return "GRN_SCORER_ERROR"
	case -77:
		return "GRN_CANCEL"
	case -78:
		return "GRN_WINDOW_FUNCTION_ERROR"
	case -79:
		return "GRN_ZSTD_ERROR"

	case InvalidAddress:
		return "invalid address"
	case InvalidCommand:
		return "invalid command"
	case InvalidOperation:
		return "invalid operation"
	case InvalidResponse:
		return "invalid response"
	case InvalidType:
		return "invalid type"
	case NetworkError:
		return "network error"
	case UnknownError:
		return "unknown error"

	default:
		if text := http.StatusText(code); text != "" {
			return text
		}
		return "undefined error"
	}
}

// Error stores an error.
type Error struct {
	Code int                    `json:"code"`
	Text string                 `json:"text"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// NewError returns a new Error.
func NewError(code int, data map[string]interface{}) error {
	err := &Error{
		Code: code,
		Text: getCodeText(code),
		Data: make(map[string]interface{}),
	}
	for k, v := range data {
		err.Data[k] = v
	}
	return err
}

// EnhanceError adds data to err and returns the modified Error.
func EnhanceError(err error, data map[string]interface{}) error {
	if e, ok := err.(*Error); ok {
		for k, v := range data {
			e.Data[k] = v
		}
		return e
	}
	e := NewError(UnknownError, data).(*Error)
	if _, ok := e.Data["error"]; !ok {
		data["error"] = err.Error()
	}
	return e
}

// Error returns the JSON-encoded error object.
func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}