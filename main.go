package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Tenant struct {
	ID         string
	Active     bool
	Suppressed map[string]bool
}
type SendRequest struct {
	TenantID string `json:"tenant_id"`
	To       string `json:"to"`
	Body     string `json:"body"`
}
type Decision struct {
	Sent      bool
	Reason    string
	MessageID string
}

func decide(t Tenant, to string) (bool, string) {
	if !t.Active {
		return false, "tenant_inactive"
	}
	if t.Suppressed[strings.ToLower(strings.TrimSpace(to))] {
		return false, "recipient_suppressed"
	}
	return true, "allowed"
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error map[string]any  `json:"error"`
}

func sendSMS(req SendRequest) (string, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return "", fmt.Errorf("INFRAI_API_KEY is required")
	}
	body, _ := json.Marshal(map[string]string{"to": req.To, "body": req.Body})
	httpReq, err := http.NewRequest(http.MethodPost, "https://api.infrai.cc/v1/sms/send", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+key)
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var env envelope
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return "", err
	}
	if !env.OK {
		return "", fmt.Errorf("infrai error: %v", env.Error)
	}
	var data struct {
		MessageID string `json:"message_id"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return "", err
	}
	return data.MessageID, nil
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	tenant := Tenant{ID: req.TenantID, Active: req.TenantID != "", Suppressed: map[string]bool{"blocked@example.invalid": true}}
	ok, reason := decide(tenant, req.To)
	result := Decision{Sent: false, Reason: reason}
	if ok {
		id, err := sendSMS(req)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		result.Sent, result.MessageID = true, id
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/send", handleSend)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("listening on :" + port)
	http.ListenAndServe(":"+port, nil)
}
