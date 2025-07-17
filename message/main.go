package message

import "github.com/gin-gonic/gin"

var Http = map[int]string{
    200: "Ok",
    201: "Created",
    204: "No content",
    301: "Moved permanently",
    307: "Temporary Redirect",
    308: "Permanent Redirect",
    400: "Bad Request",
    401: "Unauthorized",
    403: "Forbidden",
    404: "Not Found",
    405: "Method Not Allowed",
    406: "Not Acceptable",
    409: "Conflict",
    429: "Too many requests",
    500: "Internal Server Error",
    501: "Not Implemented",
    502: "Bad Gateway",
    503: "Service Unavailable",
}

func ReturnMessage(code int) any {
    return gin.H{
        "message": Http[code],
    }
}

func ReturnCustomMessage(msg string) gin.H {
    return gin.H{
        "message": msg,
    }
}

func ReturnCustomDataWithKey(field string, value interface{}) gin.H {
    return gin.H{
        field: value,
    }
}

func ReturnCustomDataWithoutKey(data map[string]interface{}) gin.H {
    result := gin.H{}
    for key, value := range data {
        result[key] = value
    }
    return result
}

func ReturnSomethingWentWrongMsg() gin.H {
    return gin.H{
        "message": "something went wrong",
    }
}

func ReturnInvalidFieldMsg() gin.H {
    return gin.H{
        "message": "invalid field type",
    }
}
