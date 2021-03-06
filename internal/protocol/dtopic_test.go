// Copyright 2018-2021 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDTopicMessage_Encode(t *testing.T) {
	buf := new(bytes.Buffer)
	msg := NewDTopicMessage(OpDTopicPublish)
	msg.SetBuffer(buf)
	msg.SetDTopic("mydtopic")
	msg.SetValue([]byte("myvalue"))
	err := msg.Encode()
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}
}

func TestDTopicMessage_Decode(t *testing.T) {
	buf := new(bytes.Buffer)

	// Encode first
	msg := NewDTopicMessage(OpStreamMessage)
	msg.SetBuffer(buf)
	msg.SetDTopic("mydtopic")
	msg.SetValue([]byte("myvalue"))
	msg.SetExtra(StreamMessageExtra{
		ListenerID: 10,
	})

	err := msg.Encode()
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}

	conn := newFakeTCPConn(buf.Bytes())
	buf.Reset()
	header, err := ReadMessage(conn, buf)
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}
	if header.Magic != MagicDTopicReq {
		t.Fatalf("Expected MagicDTopicReq (%d). Got: %d", MagicDTopicReq, header.Magic)
	}

	// Decode message from the TCP socket
	req := NewDTopicMessageFromRequest(buf)
	err = req.Decode()
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}

	if !bytes.Equal(req.Value(), []byte("myvalue")) {
		t.Fatalf("Expected myvalue. Got: %v", string(req.Value()))
	}

	if !reflect.DeepEqual(msg.Extra(), req.Extra()) {
		t.Fatalf("Different extra")
	}
}

func TestNewDTopicMessage_Response(t *testing.T) {
	buf := new(bytes.Buffer)
	msg := NewDTopicMessage(OpDTopicPublish)
	msg.SetBuffer(buf)

	err := msg.Encode()
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}

	respBuf := new(bytes.Buffer)
	resp := msg.Response(respBuf)
	if resp.OpCode() != msg.OpCode() {
		t.Fatalf("Expected OpCode: %d. Got: %d", msg.OpCode(), resp.OpCode())
	}

	value := []byte("value")
	resp.SetValue(value)
	if !bytes.Equal(resp.Value(), value) {
		t.Fatalf("response.Value() returned a different value")
	}

	resp.SetStatus(StatusErrInternalFailure)
	if resp.Status() != StatusErrInternalFailure {
		t.Fatalf("Expected status code: %d. Got: %d", StatusErrInternalFailure, resp.Status())
	}
}
