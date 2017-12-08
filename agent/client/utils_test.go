package client_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/sdk/agent"
)

type mockHTTPServer struct{}

func (m *mockHTTPServer) sendResponse(w http.ResponseWriter, data interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (m *mockHTTPServer) mockGetInfo(w http.ResponseWriter, r *http.Request) {
	p := agent.Info{}
	m.sendResponse(w, p)
}

func mockAgentHTTPServer(t *testing.T, address string) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObj := agent.NewMockAgent(mockCtrl)
	mockObj.EXPECT().HttpServer().DoAndReturn(func() *http.Server {
		mux := http.NewServeMux()
		handler := mockHTTPServer{}
		mux.HandleFunc("/", handler.mockGetInfo)
		return &http.Server{
			Addr:    ":3000",
			Handler: mux,
		}
	}).Times(1)

	server := mockObj.HttpServer()
	go server.ListenAndServe()
}
