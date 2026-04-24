protocol AliasData {
    id: 0x7000
    
    alias ID: uint32
    alias Name: bytes
    
    user_id: ID
    name_len: uint16
    username: Name length_from: name_len
}
