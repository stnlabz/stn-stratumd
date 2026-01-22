package core

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type rpcRequest struct {
    JSONRPC string        `json:"jsonrpc"`
    Method  string        `json:"method"`
    Params  []interface{} `json:"params"`
    ID      int           `json:"id"`
}
type rpcResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    Result  json.RawMessage `json:"result"`
    Error   interface{}     `json:"error"`
    ID      int             `json:"id"`
}

func (c *RPCClient) GetWork() (map[string]interface{}, error) {
    reqObj := rpcRequest{JSONRPC: "2.0", Method: "getwork", Params: []interface{}{}, ID: 1}
    body, _ := json.Marshal(reqObj)
    resp, err := http.Post(c.Endpoint, "application/json", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var rpcResp rpcResponse
    if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
        return nil, err
    }
    var work map[string]interface{}
    if err := json.Unmarshal(rpcResp.Result, &work); err != nil {
        return nil, err
    }
    return work, nil
}

func (c *RPCClient) SubmitShare(jobID, nonce, ntime string) (bool, error) {
    reqObj := rpcRequest{JSONRPC: "2.0", Method: "submit", Params: []interface{}{jobID, nonce, ntime}, ID: 2}
    body, _ := json.Marshal(reqObj)
    resp, err := http.Post(c.Endpoint, "application/json", bytes.NewReader(body))
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    var rpcResp rpcResponse
    if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
        return false, err
    }
    var ok bool
    if err := json.Unmarshal(rpcResp.Result, &ok); err != nil {
        return false, err
    }
    return ok, nil
}
