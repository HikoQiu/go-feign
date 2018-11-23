package feign

import (
    "reflect"
    "regexp"
    "net/http"
    "encoding/json"
    "encoding/xml"
    "strings"
    "bytes"
    "sync"
    "path"
)

const (
    plainTextType   = "text/plain; charset=utf-8"
    jsonContentType = "application/json; charset=utf-8"
    formContentType = "application/x-www-form-urlencoded"
)

var (
    hdrUserAgentKey     = http.CanonicalHeaderKey("User-Agent")
    hdrAcceptKey        = http.CanonicalHeaderKey("Accept")
    hdrContentTypeKey   = http.CanonicalHeaderKey("Content-Type")
    hdrContentLengthKey = http.CanonicalHeaderKey("Content-Length")

    jsonCheck = regexp.MustCompile(`(?i:(application|text)/(problem\+json|json))`)
    xmlCheck  = regexp.MustCompile(`(?i:(application|text)/(problem\+xml|xml))`)

    bufPool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

func kindOf(v interface{}) reflect.Kind {
    return typeOf(v).Kind()
}

func typeOf(i interface{}) reflect.Type {
    return indirect(valueOf(i)).Type()
}

func valueOf(i interface{}) reflect.Value {
    return reflect.ValueOf(i)
}

func indirect(v reflect.Value) reflect.Value {
    return reflect.Indirect(v)
}

// detect content type by request body
func DetectContentType(body interface{}) string {
    contentType := plainTextType
    kind := kindOf(body)
    switch kind {
    case reflect.Struct, reflect.Map:
        contentType = jsonContentType
    case reflect.String:
        contentType = plainTextType
    default:
        if b, ok := body.([]byte); ok {
            contentType = http.DetectContentType(b)
        } else if kind == reflect.Slice {
            contentType = jsonContentType
        }
    }

    return contentType
}

// IsJSONType method is to check JSON content type or not
func IsJSONType(ct string) bool {
    return jsonCheck.MatchString(ct)
}

// IsXMLType method is to check XML content type or not
func IsXMLType(ct string) bool {
    return xmlCheck.MatchString(ct)
}

// Unmarshal content into object from JSON or XML
// Deprecated: kept for backward compatibility
func Unmarshal(ct string, b []byte, d interface{}) (err error) {
    if IsJSONType(ct) {
        err = json.Unmarshal(b, d)
    } else if IsXMLType(ct) {
        err = xml.Unmarshal(b, d)
    }

    return
}

func getPointer(v interface{}) interface{} {
    vv := valueOf(v)
    if vv.Kind() == reflect.Ptr {
        return v
    }
    return reflect.New(vv.Type()).Interface()
}

// IsStringEmpty method tells whether given string is empty or not
func IsStringEmpty(str string) bool {
    return len(strings.TrimSpace(str)) == 0
}

func acquireBuffer() *bytes.Buffer {
    return bufPool.Get().(*bytes.Buffer)
}

func releaseBuffer(buf *bytes.Buffer) {
    if buf != nil {
        buf.Reset()
        bufPool.Put(buf)
    }
}

func composeRequestURL(pathURL string, c *Lbc, r *LbcReq) string {
    if !strings.HasPrefix(pathURL, "/") {
        pathURL = "/" + pathURL
    }

    hasTrailingSlash := false
    if strings.HasSuffix(pathURL, "/") && len(pathURL) > 1 {
        hasTrailingSlash = true
    }

    reqURL := "/"
    for _, segment := range strings.Split(pathURL, "/") {
        if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
            key := segment[1 : len(segment)-1]
            if val, found := r.pathParams[key]; found {
                reqURL = path.Join(reqURL, val)
                continue
            }

            if val, found := r.pathParams[key]; found {
                reqURL = path.Join(reqURL, val)
                continue
            }
        }

        reqURL = path.Join(reqURL, segment)
    }

    if hasTrailingSlash {
        reqURL = reqURL + "/"
    }

    return reqURL
}

func firstNonEmpty(v ...string) string {
    for _, s := range v {
        if !IsStringEmpty(s) {
            return s
        }
    }
    return ""
}
