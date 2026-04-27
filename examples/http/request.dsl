protocol HTTPRequest {
    id: 0x5000
    method: enum {
        GET = 0
        POST = 1
        PUT = 2
        DELETE = 3
    }
    url_len: uint16
    url: bytes length_from: url_len
    headers_len: uint16
    headers: bytes length_from: headers_len
    body_len: uint32
    body: bytes length_from: body_len if method == POST || method == PUT
}
