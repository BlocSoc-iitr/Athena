package main

import (
    "bytes"
    "archive/zip"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "net/url"
)

var (
    classHash  string
    jsonRpcUrl string
)

func init() {
    flag.StringVar(&classHash, "classHash", "", "The contract class hash")
    flag.StringVar(&jsonRpcUrl, "jsonRpcUrl", "", "The JSON-RPC URL")
}

func enableCors(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func createZipFile(files []string) ([]byte, error) {
    var buf bytes.Buffer
    zipWriter := zip.NewWriter(&buf)

    for _, file := range files {
        fileToZip, err := os.Open(file)
        if err != nil {
            return nil, fmt.Errorf("failed to open file %s: %v", file, err)
        }
        defer fileToZip.Close()

        w, err := zipWriter.Create(filepath.Base(file))
        if err != nil {
            return nil, fmt.Errorf("failed to create zip entry for file %s: %v", file, err)
        }

        if _, err := io.Copy(w, fileToZip); err != nil {
            return nil, fmt.Errorf("failed to write file %s to zip: %v", file, err)
        }
    }

    if err := zipWriter.Close(); err != nil {
        return nil, fmt.Errorf("failed to close zip writer: %v", err)
    }

    return buf.Bytes(), nil
}

func GetABIHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Received request to fetch ABI")
    enableCors(w)

    var requestData map[string]string
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        log.Printf("Error decoding request body: %v", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    classHash := requestData["classHash"]
    jsonRpcUrl := requestData["jsonRpcUrl"]

    log.Printf("classHash: %s, jsonRpcUrl: %s", classHash, jsonRpcUrl)

    abi, err := GetStarknetABI(classHash, jsonRpcUrl)
    if err != nil {
        log.Printf("Error fetching ABI: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    log.Println("Successfully fetched ABI")
    w.Header().Set("Content-Type", "application/json")
    w.Write(abi)
}

func GetStarknetABI(classHash, jsonRpcUrl string) ([]byte, error) {
    log.Printf("Running CLI command to fetch ABI with classHash: %s, jsonRpcUrl: %s", classHash, jsonRpcUrl)

    cmd := exec.Command("go", "run", "cli/get/starknet.go", "--classHash", classHash, "--jsonRpcUrl", jsonRpcUrl, "--output", "abi.json")
    var out, stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        log.Printf("Error running CLI tool: %v, stderr: %s", err, stderr.String())
        return nil, fmt.Errorf("error running CLI tool: %v, stderr: %s", err, stderr.String())
    }

    log.Println("CLI command executed successfully")

    abiJson, err := os.ReadFile("abi.json")
    if err != nil {
        log.Printf("Error reading ABI file: %v", err)
        return nil, fmt.Errorf("error reading ABI file: %v", err)
    }

    log.Println("ABI file read successfully")
    return abiJson, nil
}
func GetBackfillHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Received request to backfill block data")
    enableCors(w)

    var requestData map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        log.Printf("Error decoding request body: %v", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    fromBlock := uint64(requestData["fromBlock"].(float64))
    toBlock := uint64(requestData["toBlock"].(float64))
    encodedRpcUrl := requestData["rpcUrl"].(string)
    outputFile := requestData["outputFile"].(string)
    transactionHashFlag := requestData["transactionHashFlag"].(bool)

    rpcUrl, err := url.QueryUnescape(encodedRpcUrl)
    if err != nil {
        log.Printf("Error decoding RPC URL: %v", err)
        http.Error(w, fmt.Sprintf("Error decoding RPC URL: %v", err), http.StatusInternalServerError)
        return
    }

    outputDir := filepath.Dir(outputFile)
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        log.Printf("Error creating output directory: %v", err)
        http.Error(w, fmt.Sprintf("Error creating output directory: %v", err), http.StatusInternalServerError)
        return
    }

    if err := RunBackfillCommand(fromBlock, toBlock, rpcUrl, outputFile, transactionHashFlag); err != nil {
        log.Printf("Error running backfill: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    files := []string{outputFile}
    if transactionHashFlag {
        transactionFile := "transaction_hashes_block_details.csv"
        if _, err := os.Stat(transactionFile); os.IsNotExist(err) {
            log.Printf("Expected transaction hash file not found: %v", err)
            http.Error(w, fmt.Sprintf("Expected file not found: %s", transactionFile), http.StatusInternalServerError)
            return
        }
        files = append(files, transactionFile)
    }

    zipData, err := createZipFile(files)
    if err != nil {
        log.Printf("Error creating zip file: %v", err)
        http.Error(w, fmt.Sprintf("Error creating zip file: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", "attachment; filename=backfill_results.zip")
    _, err = w.Write(zipData)
    if err != nil {
        log.Printf("Error writing ZIP file to response: %v", err)
        http.Error(w, "Error writing ZIP file to response", http.StatusInternalServerError)
    }
}


func RunBackfillCommand(fromBlock, toBlock uint64, rpcUrl, outputFile string, transactionHashFlag bool) error {
    args := []string{
        "run", "cli/backfill/starknet.go",
        "--from", fmt.Sprintf("%d", fromBlock),
        "--to", fmt.Sprintf("%d", toBlock),
        "--rpc-url", rpcUrl,
        "--output", outputFile,
    }

    if transactionHashFlag {
        args = append(args, "--transactionhash")
    }

    cmd := exec.Command("go", args...)
    var out, stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        log.Printf("Error running backfill command: %v", err)
        log.Printf("Stderr output: %s", stderr.String())
        return fmt.Errorf("command failed: %v, stderr: %s", err, stderr.String())
    }

    log.Println("Backfill command executed successfully")
    log.Printf("Command output: %s", out.String())
    return nil
}

func FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
    enableCors(w)

    filePath := r.URL.Query().Get("file")
    if filePath == "" {
        http.Error(w, "File path is required", http.StatusBadRequest)
        return
    }

    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    http.ServeFile(w, r, filePath)
}

func main() {
    flag.Parse()

    if classHash != "" && jsonRpcUrl != "" {
        abi, err := GetStarknetABI(classHash, jsonRpcUrl)
        if err != nil {
            log.Fatalf("Error fetching ABI: %v", err)
        }
        fmt.Println(string(abi))
        return
    }

    http.HandleFunc("/api/abi", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "OPTIONS" {
            enableCors(w)
            return
        }
        GetABIHandler(w, r)
    })

    http.HandleFunc("/api/backfill", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "OPTIONS" {
            enableCors(w)
            return
        }
        GetBackfillHandler(w, r)
    })

    http.HandleFunc("/api/download", FileDownloadHandler)

    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
