package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mvanhorn/cars-il-pp-cli/internal/client"
	"github.com/mvanhorn/cars-il-pp-cli/internal/cli"
	"github.com/mvanhorn/cars-il-pp-cli/internal/mcp"
	"github.com/mvanhorn/cars-il-pp-cli/internal/store"
)

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

var version = "1.0.0"

func main() {
	db, err := store.Open("")
	if err != nil {
		write(response{JSONRPC: "2.0", Error: map[string]interface{}{"code": -32000, "message": err.Error()}})
		os.Exit(1)
	}
	defer db.Close()
	service := commands.NewService(db)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			write(response{JSONRPC: "2.0", Error: map[string]interface{}{"code": -32700, "message": "parse error"}})
			continue
		}
		result, callErr := handle(service, req.Method, req.Params)
		resp := response{JSONRPC: "2.0", ID: req.ID}
		if callErr != nil {
			resp.Error = map[string]interface{}{"code": -32000, "message": callErr.Error()}
		} else {
			resp.Result = result
		}
		write(resp)
	}
}

func handle(service commands.Service, method string, params json.RawMessage) (interface{}, error) {
	switch method {
	case "initialize":
		return map[string]interface{}{"protocolVersion": "2024-11-05", "serverInfo": map[string]string{"name": "cars-il-pp-mcp", "version": version}}, nil
	case "tools/list":
		return map[string]interface{}{"tools": mcp.RegisterTools()}, nil
	case "tools/call":
		var call struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(params, &call); err != nil {
			return nil, err
		}
		return callTool(service, call.Name, call.Arguments)
	default:
		return nil, fmt.Errorf("unsupported MCP method: %s", method)
	}
}

func callTool(service commands.Service, name string, raw json.RawMessage) (interface{}, error) {
	var params client.SearchParams
	_ = json.Unmarshal(raw, &params)
	switch name {
	case "search":
		return service.Search(params)
	case "listings_get":
		var in struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(raw, &in)
		return service.Get(in.ID, params.DataSource)
	case "sync":
		return service.Sync(params)
	case "market":
		return service.Market(params)
	case "deal":
		var in struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(raw, &in)
		return service.Deal(in.ID)
	case "price_history":
		var in struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(raw, &in)
		return service.DB.GetPriceHistory(in.ID)
	case "stale", "compare", "depreciation", "market_heat", "watch", "doctor":
		return map[string]string{"message": "operation is available in the CLI; use matching cars-il command for streaming/table behavior"}, nil
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func write(resp response) {
	_ = json.NewEncoder(os.Stdout).Encode(resp)
}
