package core

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "time"
)

func Run() {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
    if err != nil { log.Fatalf("listen error: %v", err) }
    log.Printf("Stratum listening on :%d", cfg.Port)
    for {
        c, err := ln.Accept()
        if err != nil { log.Println("accept:", err); continue }
        go handleConn(c)
    }
}

func handleConn(conn net.Conn) {
    defer conn.Close()
    rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

    // subscribe + authorize
    writeJSON(rw, map[string]interface{}{"id": 1, "method": "mining.subscribe", "params": []interface{}{"stratumd/clean"}})
    writeJSON(rw, map[string]interface{}{"id": 2, "method": "mining.authorize", "params": []interface{}{"", ""}})

    // announce immediately + on interval
    announce(rw)
    t := time.NewTicker(time.Duration(cfg.WorkInterval) * time.Second)
    defer t.Stop()
    go func() {
        for range t.C { announce(rw) }
    }()

    // read share submissions
    for {
        var req struct {
            ID     interface{}   `json:"id"`
            Method string        `json:"method"`
            Params []interface{} `json:"params"`
        }
        if err := json.NewDecoder(rw).Decode(&req); err != nil {
            log.Println("read:", err)
            return
        }
        if req.Method != "mining.submit" {
            continue
        }
        if len(req.Params) < 4 {
            writeJSON(rw, map[string]interface{}{"id": req.ID, "result": false, "error": "invalid params"})
            continue
        }
        // [user, jobID, nonce, ntime]
        jobID, _ := req.Params[1].(string)
        nonce, _ := req.Params[2].(string)
        ntime, _ := req.Params[3].(string)

        ok, err := rpcClient.SubmitShare(jobID, nonce, ntime)
        var e interface{}
        if err != nil { e = err.Error() }
        writeJSON(rw, map[string]interface{}{"id": req.ID, "result": ok, "error": e})
    }
}

func announce(rw *bufio.ReadWriter) {
    work, err := rpcClient.GetWork()
    if err != nil {
        log.Println("GetWork:", err)
        return
    }
    params := buildParams(work) // always 9

    nbits := fmt.Sprintf("%v", params[6])
    diff := DifficultyFromNBits(nbits)

    log.Printf(
        "Announcing job %v (nbits=%v ntime=%v diff=%s)",
        params[0], nbits, params[7], HumanDiff(diff),
    )

    writeJSON(rw, map[string]interface{}{
        "id":     nil,
        "method": "mining.notify",
        "params": params,
    })
}

// translate your minimal getwork -> full 9-param Stratum notify
// [job_id, prevhash, coinb1, coinb2, merkle_branches, version, nbits, ntime, clean_jobs]
func buildParams(work map[string]interface{}) []interface{} {
    prev   := strOr(work["prevhash"], "00")
    ntime  := strOr(work["ntime"],   "00000000")
    nbits  := strOr(work["target"],  "1d00ffff") // treat your 'target' as compact bits

    jobID    := fmt.Sprintf("job-%d", time.Now().UnixNano())
    coinb1   := ""                // placeholders until your node provides them
    coinb2   := ""
    branches := []interface{}{}
    version  := "20000000"        // 0x20000000 (little endian) as hex string
    clean    := true

    return []interface{}{jobID, prev, coinb1, coinb2, branches, version, nbits, ntime, clean}
}

func strOr(v interface{}, def string) string {
    if s, ok := v.(string); ok && s != "" {
        return s
    }
    return def
}

func writeJSON(rw *bufio.ReadWriter, v interface{}) {
    if err := json.NewEncoder(rw).Encode(v); err != nil {
        log.Println("encode:", err); return
    }
    rw.Flush()
}
