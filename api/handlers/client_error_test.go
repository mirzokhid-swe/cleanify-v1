package handlers

import (
	"bw-erp/models"
	"testing"
	"time"
)

func TestClientErrorAllowed(t *testing.T) {
	clientErrorMinute, clientErrorCount = 0, 0
	now := time.Unix(1000*60+5, 0)
	for i := 0; i < models.ClientErrorMaxPerMinute; i++ {
		if !clientErrorAllowed(now) {
			t.Fatalf("%d-xabar o'tishi kerak", i+1)
		}
	}
	if clientErrorAllowed(now.Add(10 * time.Second)) {
		t.Fatal("shu daqiqada chegaradan oshgan xabar o'tmasligi kerak")
	}
	if !clientErrorAllowed(now.Add(60 * time.Second)) {
		t.Fatal("keyingi daqiqada yana o'tishi kerak")
	}
}
