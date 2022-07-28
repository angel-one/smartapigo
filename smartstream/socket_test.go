package smartstream

import "testing"

func TestSmartStream(t *testing.T) {
	client := New("A586457", "00998877")
	client.Connect()
}
