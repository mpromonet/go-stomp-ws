/*
 * SPDX-License-Identifier: Unlicense
 *
 * This is free and unencumbered software released into the public domain.
 *
 * Anyone is free to copy, modify, publish, use, compile, sell, or distribute this
 * software, either in source code form or as a compiled binary, for any purpose,
 * commercial or non-commercial, and by any means.
 *
 * For more information, please refer to <http://unlicense.org/>
 */

package cmd

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

	"github.com/go-stomp/stomp/v3/frame"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

func sendFrame(conn *websocket.Conn, f *frame.Frame) {
	var buf bytes.Buffer
	writer := frame.NewWriter(&buf)
	writer.Write(f)
	fmt.Printf("Sending frame: %v\n", string(buf.Bytes()))
	err := conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}
}

func getFrame(conn *websocket.Conn) (*frame.Frame, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Recv frame: %v\n", string(message))
	reader := frame.NewReader(bytes.NewReader(message))
	return reader.Read()
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Connect to a WS server",
	Run: func(cmd *cobra.Command, args []string) {
		wsurl, _ := cmd.Flags().GetString("url")
		topic, _ := cmd.Flags().GetString("topic")
		message, _ := cmd.Flags().GetString("message")

		conn, _, err := websocket.DefaultDialer.Dial(wsurl, nil)
		if err != nil {
			log.Fatalf("failed to connect to WebSocket server: %v", err)
		}
		defer conn.Close()

		connectFrame := frame.New(frame.CONNECT, frame.AcceptVersion, "1.1")
		sendFrame(conn, connectFrame)

		getFrame(conn)

		if message == "" {
			subFrame := frame.New(frame.SUBSCRIBE, frame.Destination, topic, frame.Id, "0", frame.Ack, "auto")
			sendFrame(conn, subFrame)
			for {
				_, err := getFrame(conn)
				if err != nil {
					log.Printf("Received message: error %v", err)
					break
				}
			}
		} else {
			body := []byte(message)
			pubFrame := frame.New(frame.SEND, frame.Destination, topic, frame.ContentLength, strconv.Itoa(len(body)))
			pubFrame.Body = body
			sendFrame(conn, pubFrame)

			getFrame(conn)
		}
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringP("url", "u", "ws://localhost:8765/ws", "ws server url to connect")
	clientCmd.Flags().StringP("topic", "t", "/topic/notifications", "topic to use")
	clientCmd.Flags().StringP("message", "m", "", "message to send (empty to subscribe)")
}
