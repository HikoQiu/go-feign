package feign

import (
    "net/url"
    "net/http"
    "strings"
    "bytes"
)

// Load balance request entity
type LbcReq struct {
    // Load balance client
    lbc *Lbc

    // Load balance request params
    URL    string
    Method string

    QueryParam url.Values
    FormData   url.Values
    pathParams map[string]string

    Header http.Header
    Body   interface{}

    // vars to accept response by auto unmashal
    Result interface{}
    Error  interface{}

    //Time       time.Time
    //RawRequest *http.Request

    isMultiPart bool
    isFormData  bool

    // parsed request body bytes
    bodyBuf          *bytes.Buffer
    setContentLength bool
    //isSaveResponse   bool
    //outputFile       string
    //notParseResponse bool
    //ctx                 context.Context
    fallbackContentType string
}

// Set HTTP Request params

// SetHeader method is to set a single header field and its value in the current request.
// Example: To set `Content-Type` and `Accept` as `application/json`.
// 		resty.R().
//			SetHeader("Content-Type", "application/json").
//			SetHeader("Accept", "application/json")
//
// Also you can override header value, which was set at client instance level.
//
func (r *LbcReq) SetHeader(header, value string) *LbcReq {
    r.Header.Set(header, value)
    return r
}

// SetHeaders method sets multiple headers field and its values at one go in the current request.
// Example: To set `Content-Type` and `Accept` as `application/json`
//
// 		resty.R().
//			SetHeaders(map[string]string{
//				"Content-Type": "application/json",
//				"Accept": "application/json",
//			})
// Also you can override header value, which was set at client instance level.
//
func (r *LbcReq) SetHeaders(headers map[string]string) *LbcReq {
    for h, v := range headers {
        r.SetHeader(h, v)
    }

    return r
}

// SetQueryParam method sets single parameter and its value in the current request.
// It will be formed as query string for the request.
// Example: `search=kitchen%20papers&size=large` in the URL after `?` mark.
// 		resty.R().
//			SetQueryParam("search", "kitchen papers").
//			SetQueryParam("size", "large")
// Also you can override query params value, which was set at client instance level
//
func (r *LbcReq) SetQueryParam(param, value string) *LbcReq {
    r.QueryParam.Set(param, value)
    return r
}

// SetQueryParams method sets multiple parameters and its values at one go in the current request.
// It will be formed as query string for the request.
// Example: `search=kitchen%20papers&size=large` in the URL after `?` mark.
// 		resty.R().
//			SetQueryParams(map[string]string{
//				"search": "kitchen papers",
//				"size": "large",
//			})
// Also you can override query params value, which was set at client instance level
//
func (r *LbcReq) SetQueryParams(params map[string]string) *LbcReq {
    for p, v := range params {
        r.SetQueryParam(p, v)
    }

    return r
}

// SetMultiValueQueryParams method appends multiple parameters with multi-value
// at one go in the current request. It will be formed as query string for the request.
// Example: `status=pending&status=approved&status=open` in the URL after `?` mark.
// 		resty.R().
//			SetMultiValueQueryParams(url.Values{
//				"status": []string{"pending", "approved", "open"},
//			})
// Also you can override query params value, which was set at client instance level
//
func (r *LbcReq) SetMultiValueQueryParams(params url.Values) *LbcReq {
    for p, v := range params {
        for _, pv := range v {
            r.QueryParam.Add(p, pv)
        }
    }

    return r
}

// SetQueryString method provides ability to use string as an input to set URL query string for the request.
//
// Using String as an input
// 		resty.R().
//			SetQueryString("productId=232&template=fresh-sample&cat=resty&source=google&kw=buy a lot more")
//
func (r *LbcReq) SetQueryString(query string) *LbcReq {
    params, err := url.ParseQuery(strings.TrimSpace(query))
    if err == nil {
        for p, v := range params {
            for _, pv := range v {
                r.QueryParam.Add(p, pv)
            }
        }
    } else {
        log.Errorf("ERROR [%v]", err)
    }
    return r
}

// SetFormData method sets Form parameters and their values in the current request.
// It's applicable only HTTP method `POST` and `PUT` and requests content type would be set as
// `application/x-www-form-urlencoded`.
// 		resty.R().
// 			SetFormData(map[string]string{
//				"access_token": "BC594900-518B-4F7E-AC75-BD37F019E08F",
//				"user_id": "3455454545",
//			})
// Also you can override form data value, which was set at client instance level
//
func (r *LbcReq) SetFormData(data map[string]string) *LbcReq {
    for k, v := range data {
        r.FormData.Set(k, v)
    }

    return r
}

// SetMultiValueFormData method appends multiple form parameters with multi-value
// at one go in the current request.
// 		resty.R().
//			SetMultiValueFormData(url.Values{
//				"search_criteria": []string{"book", "glass", "pencil"},
//			})
// Also you can override form data value, which was set at client instance level
//
func (r *LbcReq) SetMultiValueFormData(params url.Values) *LbcReq {
    for k, v := range params {
        for _, kv := range v {
            r.FormData.Add(k, kv)
        }
    }

    return r
}

// SetBody method sets the request body for the request. It supports various realtime need easy.
// We can say its quite handy or powerful. Supported request body data types is `string`, `[]byte`,
// `struct` and `map`. Body value can be pointer or non-pointer. Automatic marshalling
// for JSON and XML content type, if it is `struct` or `map`.
//
// Example:
//
// Struct as a body input, based on content type, it will be marshalled.
//		resty.R().
//			SetBody(User{
//				Username: "jeeva@myjeeva.com",
//				Password: "welcome2resty",
//			})
//
// Map as a body input, based on content type, it will be marshalled.
//		resty.R().
//			SetBody(map[string]interface{}{
//				"username": "jeeva@myjeeva.com",
//				"password": "welcome2resty",
//				"address": &Address{
//					Address1: "1111 This is my street",
//					Address2: "Apt 201",
//					City: "My City",
//					State: "My State",
//					ZipCode: 00000,
//				},
//			})
//
// String as a body input. Suitable for any need as a string input.
//		resty.R().
//			SetBody(`{
//				"username": "jeeva@getrightcare.com",
//				"password": "admin"
//			}`)
//
// []byte as a body input. Suitable for raw request such as file upload, serialize & deserialize, etc.
// 		resty.R().
//			SetBody([]byte("This is my raw request, sent as-is"))
//
func (r *LbcReq) SetBody(body interface{}) *LbcReq {
    r.Body = body
    return r
}

// SetResult method is to register the response `Result` object for automatic unmarshalling in the RESTful mode
// if response status code is between 200 and 299 and content type either JSON or XML.
//
// Note: Result object can be pointer or non-pointer.
//		resty.R().SetResult(&AuthToken{})
//		// OR
//		resty.R().SetResult(AuthToken{})
//
// Accessing a result value
//		response.Result().(*AuthToken)
//
func (r *LbcReq) SetResult(res interface{}) *LbcReq {
    r.Result = getPointer(res)
    return r
}

// SetError method is to register the request `Error` object for automatic unmarshalling in the RESTful mode
// if response status code is greater than 399 and content type either JSON or XML.
//
// Note: Error object can be pointer or non-pointer.
// 		resty.R().SetError(&AuthError{})
//		// OR
//		resty.R().SetError(AuthError{})
//
// Accessing a error value
//		response.Error().(*AuthError)
//
func (r *LbcReq) SetError(err interface{}) *LbcReq {
    r.Error = getPointer(err)
    return r
}

// SetPathParams method sets multiple URL path key-value pairs at one go in the
// resty current request instance.
// 		resty.R().SetPathParams(map[string]string{
// 		   "userId": "sample@sample.com",
// 		   "subAccountId": "100002",
// 		})
//
// 		Result:
// 		   URL - /v1/users/{userId}/{subAccountId}/details
// 		   Composed URL - /v1/users/sample@sample.com/100002/details
// It replace the value of the key while composing request URL. Also you can
// override Path Params value, which was set at client instance level.
func (r *LbcReq) SetPathParams(params map[string]string) *LbcReq {
    for p, v := range params {
        r.pathParams[p] = v
    }
    return r
}

// ExpectContentType method allows to provide fallback `Content-Type` for automatic unmarshalling
// when `Content-Type` response header is unavailable.
func (r *LbcReq) ExpectContentType(contentType string) *LbcReq {
    r.fallbackContentType = contentType
    return r
}

//------------------------------//
// HTTP verb method starts here //
//------------------------------//

// Get method does GET HTTP request. It's defined in section 4.3.1 of RFC7231.
func (r *LbcReq) Get(url string) (*LbcResp, error) {
    return r.Execute(http.MethodGet, url)
}

// Head method does HEAD HTTP request. It's defined in section 4.3.2 of RFC7231.
func (r *LbcReq) Head(url string) (*LbcResp, error) {
    return r.Execute(http.MethodHead, url)
}

// Post method does POST HTTP request. It's defined in section 4.3.3 of RFC7231.
func (r *LbcReq) Post(url string) (*LbcResp, error) {
    return r.Execute(http.MethodPost, url)
}

// Put method does PUT HTTP request. It's defined in section 4.3.4 of RFC7231.
func (r *LbcReq) Put(url string) (*LbcResp, error) {
    return r.Execute(http.MethodPut, url)
}

// Delete method does DELETE HTTP request. It's defined in section 4.3.5 of RFC7231.
func (r *LbcReq) Delete(url string) (*LbcResp, error) {
    return r.Execute(http.MethodDelete, url)
}

// Options method does OPTIONS HTTP request. It's defined in section 4.3.7 of RFC7231.
func (r *LbcReq) Options(url string) (*LbcResp, error) {
    return r.Execute(http.MethodOptions, url)
}

// Patch method does PATCH HTTP request. It's defined in section 2 of RFC5789.
func (r *LbcReq) Patch(url string) (*LbcResp, error) {
    return r.Execute(http.MethodPatch, url)
}

// @TODO Use LBClient to send request
// Execute method performs the HTTP request with given HTTP method and URL
// for current `Request`.
// 		resp, err := resty.R().Execute(resty.GET, "http://httpbin.org/get")
//
func (r *LbcReq) Execute(method, url string) (*LbcResp, error) {
    r.Method = method
    r.URL = url

    lbcResponse, err := r.lbc.do(r)
    if err != nil {
        log.Errorf("Failed to request, err=%s", err.Error())
        return nil, err
    }

    log.Debugf("lbc response, v=%s", lbcResponse)
    return nil, nil
}

//------------------------------//
// HTTP verb method ends here //
//------------------------------//
