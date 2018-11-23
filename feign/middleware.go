package feign

import (
    "net/http"
    "net/url"
    "bytes"
    "fmt"
    "io"
    "reflect"
    "encoding/xml"
    "errors"
    "k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/json"
    "mime/multipart"
)

func parseRequestHeader(c *Lbc, r *LbcReq) error {
    hdr := make(http.Header)
    for k := range r.Header {
        hdr.Del(k)
        hdr[k] = append(hdr[k], r.Header[k]...)
    }

    if IsStringEmpty(hdr.Get(hdrUserAgentKey)) {
        hdr.Set(hdrUserAgentKey, "go-feign v1.0")
    }

    ct := hdr.Get(hdrContentTypeKey)
    if IsStringEmpty(hdr.Get(hdrAcceptKey)) && !IsStringEmpty(ct) &&
        (IsJSONType(ct) || IsXMLType(ct)) {
        hdr.Set(hdrAcceptKey, hdr.Get(hdrContentTypeKey))
    }

    r.Header = hdr
    return nil
}

func handleFormData(c *Lbc, r *LbcReq) {
    formData := url.Values{}

    for k, v := range r.FormData {
        for _, iv := range v {
            formData.Add(k, iv)
        }
    }

    r.bodyBuf = bytes.NewBuffer([]byte(formData.Encode()))
    r.Header.Set(hdrContentTypeKey, formContentType)
    r.isFormData = true
}

func handleContentType(c *Lbc, r *LbcReq) {
    contentType := r.Header.Get(hdrContentTypeKey)
    if IsStringEmpty(contentType) {
        contentType = DetectContentType(r.Body)
        r.Header.Set(hdrContentTypeKey, contentType)
    }
}

func parseRequestBody(c *Lbc, r *LbcReq) (err error) {
    if isPayloadSupported(r.Method, c.AllowGetMethodPayload) {
        // Handling Multipart
        if r.isMultiPart && !(r.Method == http.MethodPatch) {
            if err = handleMultipart(c, r); err != nil {
                return
            }

            goto CL
        }

        // Handling Form Data
        if len(r.FormData) > 0 {
            handleFormData(c, r)

            goto CL
        }

        // Handling Request body
        if r.Body != nil {
            handleContentType(c, r)

            if err = handleRequestBody(c, r); err != nil {
                return
            }
        }
    }

CL:
// by default resty won't set content length, you can if you want to :)
    if (r.setContentLength) && r.bodyBuf != nil {
        r.Header.Set(hdrContentLengthKey, fmt.Sprintf("%d", r.bodyBuf.Len()))
    }

    return
}

func isPayloadSupported(m string, allowMethodGet bool) bool {
    return !(m == http.MethodHead || m == http.MethodOptions || (m == http.MethodGet && !allowMethodGet))
}

func handleRequestBody(c *Lbc, r *LbcReq) (err error) {
    var bodyBytes []byte
    contentType := r.Header.Get(hdrContentTypeKey)
    kind := kindOf(r.Body)
    r.bodyBuf = nil

    if reader, ok := r.Body.(io.Reader); ok {
        if r.setContentLength { // keep backward compability
            r.bodyBuf = acquireBuffer()
            _, err = r.bodyBuf.ReadFrom(reader)
            r.Body = nil
        } else {
            // Otherwise buffer less processing for `io.Reader`, sounds good.
            return
        }
    } else if b, ok := r.Body.([]byte); ok {
        bodyBytes = b
    } else if s, ok := r.Body.(string); ok {
        bodyBytes = []byte(s)
    } else if IsJSONType(contentType) &&
        (kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice) {
        bodyBytes, err = json.Marshal(r.Body)
    } else if IsXMLType(contentType) && (kind == reflect.Struct) {
        bodyBytes, err = xml.Marshal(r.Body)
    }

    if bodyBytes == nil && r.bodyBuf == nil {
        err = errors.New("unsupported 'Body' type/value")
    }

    // if any errors during body bytes handling, return it
    if err != nil {
        return
    }

    // []byte into Buffer
    if bodyBytes != nil && r.bodyBuf == nil {
        r.bodyBuf = acquireBuffer()
        _, _ = r.bodyBuf.Write(bodyBytes)
    }

    return
}

func handleMultipart(c *Lbc, r *LbcReq) (err error) {
    r.bodyBuf = acquireBuffer()
    w := multipart.NewWriter(r.bodyBuf)

    for k, v := range r.FormData {
        for _, iv := range v {
            if err = w.WriteField(k, iv); err != nil {
                return err
            }
        }
    }

    r.Header.Set(hdrContentTypeKey, w.FormDataContentType())
    err = w.Close()

    return
}

func parseRequestURL(c *Lbc, r *LbcReq) error {
    // Parsing request URL
    reqURL, err := url.Parse(r.URL)
    if err != nil {
        return err
    }

    // GitHub #103 Path Params
    reqURL.Path = composeRequestURL(reqURL.Path, c, r)

    // If Request.Url is relative path then added c.HostUrl into
    // the request URL otherwise Request.Url will be used as-is
    // @TODO url 配置上路径
    //if !reqURL.IsAbs() {
    //    reqURL, err = url.Parse(c.HostURL + reqURL.String())
    //    if err != nil {
    //        return err
    //    }
    //}

    // Adding Query Param
    query := make(url.Values)
    for k, v := range r.QueryParam {
        for _, iv := range v {
            query.Add(k, iv)
        }
    }

    for k, v := range r.QueryParam {
        // remove query param from client level by key
        // since overrides happens for that key in the request
        query.Del(k)

        for _, iv := range v {
            query.Add(k, iv)
        }
    }

    // GitHub #123 Preserve query string order partially.
    // Since not feasible in `SetQuery*` resty methods, because
    // standard package `url.Encode(...)` sorts the query params
    // alphabetically
    if len(query) > 0 {
        if IsStringEmpty(reqURL.RawQuery) {
            reqURL.RawQuery = query.Encode()
        } else {
            reqURL.RawQuery = reqURL.RawQuery + "&" + query.Encode()
        }
    }

    r.URL = reqURL.String()

    return nil
}

// @TODO 生成 HTTP Request
func createHTTPRequest(c *Lbc, r *LbcReq) (err error) {
    //if r.bodyBuf == nil {
    //    if reader, ok := r.Body.(io.Reader); ok {
    //        r.RawRequest, err = http.NewRequest(r.Method, r.URL, reader)
    //    } else {
    //        r.RawRequest, err = http.NewRequest(r.Method, r.URL, nil)
    //    }
    //} else {
    //    r.RawRequest, err = http.NewRequest(r.Method, r.URL, r.bodyBuf)
    //}
    //
    //if err != nil {
    //    return
    //}

    //// Assign close connection option
    //r.RawRequest.Close = c.closeConnection
    //
    //// Add headers into http request
    //r.RawRequest.Header = r.Header
    //
    //// Add cookies into http request
    //for _, cookie := range c.Cookies {
    //    r.RawRequest.AddCookie(cookie)
    //}
    //
    //// it's for non-http scheme option
    //if r.RawRequest.URL != nil && r.RawRequest.URL.Scheme == "" {
    //    r.RawRequest.URL.Scheme = c.scheme
    //    r.RawRequest.URL.Host = r.URL
    //}

    // Use context if it was specified
    //r.addContextIfAvailable()

    return
}

// @TODO 解析响应体
func parseResponseBody(c *Lbc, res *LbcResp) (err error) {
    // Handles only JSON or XML content type
    //ct := firstNonEmpty(res.Header().Get(hdrContentTypeKey), res.Request.fallbackContentType)
    //if IsJSONType(ct) || IsXMLType(ct) {
    //    // Considered as Result
    //    if res.StatusCode() > 199 && res.StatusCode() < 300 {
    //        if res.Request.Result != nil {
    //            err = Unmarshalc(c, ct, res.body, res.Request.Result)
    //            return
    //        }
    //    }
    //
    //    // Considered as Error
    //    if res.StatusCode() > 399 {
    //        // global error interface
    //        if res.Request.Error == nil && c.Error != nil {
    //            res.Request.Error = reflect.New(c.Error).Interface()
    //        }
    //
    //        if res.Request.Error != nil {
    //            err = Unmarshalc(c, ct, res.body, res.Request.Error)
    //        }
    //    }
    //}

    return
}
