package burst

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/burst/php"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net"
	"net/http"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This is a module code that runs burst containers.
// The big difference between these and other modules is that it actually does not have
// an entry point.

func SetupBurstLambdaEndpoint(path string, paid bool) {
	http.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		if paid {
			apiKey := request.URL.Query().Get("apikey")
			_, call := BurstSession[apiKey]
			if !call {
				management.QuantumGradeAuthorization()
				writer.WriteHeader(http.StatusPaymentRequired)
				writer.Write([]byte("Payment required with a PUT to /run.coin"))
			}
		}
		input := drawing.NoErrorString(io.ReadAll(request.Body))
		deferredKey, output := RunBurst(input)
		if deferredKey != "" && output == "" {
			time.Sleep(3 * time.Second)
			output = GetBurst(deferredKey)
		}
		drawing.NoErrorWrite64(io.Copy(writer, bytes.NewBuffer([]byte(output))))
	})
}

func SetupBurstIdleProcess() {
	go func() {
		err := acceptMessage(ProcessBoxMessageEnglang)
		if err != nil {
			//fmt.Println(err)
		}
	}()
}

func acceptMessage(handler func(string) string) error {
	// Create a UDP address to listen on
	address, err := net.ResolveUDPAddr("udp", metadata.UdpContainerPort)
	if err != nil {
		return fmt.Errorf("error10 %s", err)
	}

	// Create a UDP connection
	conn, err := net.ListenUDP("udp", address)
	if err != nil {
		return fmt.Errorf("error11 %s", err)
	}
	cleanup := func() { _ = conn.Close() }
	CleanupNetworkResources = append(CleanupNetworkResources, cleanup)
	defer cleanup()

	fmt.Println("Listening on port ", address)

	buffer := make([]byte, 1024)
	var message string
	for {
		// Read data from the connection into the buffer
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			drawing.NoErrorVoid(conn.Close())
			break
		}

		// Process the received data
		data := string(buffer[:n])
		message = message + data
		if IsMessageComplete(message) {
			reply := handler(message)
			message = ""
			if reply != "" {
				n, err = conn.WriteToUDP([]byte(reply), addr)
				if err != nil {
					break
				}
			}
		}
	}
	return nil
}

func SendMessage(address string, message string) (string, error) {
	// Create a UDP address for the destination
	udp, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return "", fmt.Errorf("error1 %s", err)
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, udp)
	if err != nil {
		return "", fmt.Errorf("error2 %s", err)
	}
	cleanup := func() { _ = conn.Close() }
	CleanupNetworkResources = append(CleanupNetworkResources, cleanup)
	defer cleanup()

	_, err = conn.Write([]byte(message))
	if err != nil {
		return "", fmt.Errorf("error3 %s", err)
	}

	buffer := make([]byte, 1024)
	var reply string
	for {
		// Read data from the connection into the buffer
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return "", fmt.Errorf("error4 %s", err)
		}

		// Process the received data
		data := string(buffer[:n])
		reply = reply + data
		if IsMessageComplete(reply) {
			return reply, nil
		} else {
			fmt.Printf("Received %d bytes from %s: %s\n", n, addr.String(), data)
		}
	}
}

func IsMessageComplete(s string) bool {
	var message string
	err := englang.Scanf1(s, "Message has started.%sMessage has finished.", &message)
	return err == nil
}

func RunExternalShell(task string) string {
	var ret string
	ret = php.EnglangPhp(drawing.GenerateUniqueKey(), task, MaxBurstRuntime)
	if ret != "" {
		return ret
	}
	return "This is the result." + task
}

func StartBurst(code string) string {
	lock.Lock()
	defer lock.Unlock()
	for containerKey, containerContent := range ContainerRunning {
		var status, key string
		if nil == englang.Scanf1(containerContent, "Burst container has key %s and it is running %s.", &key, &status) {
			if status == "idle" {
				update := englang.Printf("Burst container has key %s and it is running %s.Run this.%s", containerKey, "code", code)
				ContainerRunning[containerKey] = update
				return containerKey
			}
		}
	}
	return ""
}

func GetBurst(key string) string {
	lock.Lock()
	defer lock.Unlock()
	content, ok := ContainerRunning[key]
	if ok {
		var containerKey, result string
		if nil == englang.Scanf1(content+"DFFSSFFGGG", "Burst container has key %s and it is finished with the following result %s"+"DFFSSFFGGG", &containerKey, &result) {
			update := englang.Printf("Burst container has key %s and it is running %s.", containerKey, "idle")
			ContainerRunning[containerKey] = update
			return result
		}
	}
	return ""
}

func RunBurst(code string) (string, string) {
	containerKey := StartBurst(code)
	if containerKey == "" {
		// System is busy
		return "The system is busy. Please reload.", ""
	}
	for i := 0; i < 15; i++ {
		result := GetBurst(containerKey)
		if result != "" {
			return containerKey, result
		}
		time.Sleep(1 * time.Millisecond)
	}
	return containerKey, ""
}

func FinishCleanup() {
	ContainerRunning = map[string]string{}
	for _, v := range CleanupNetworkResources {
		v()
	}
	CleanupNetworkResources = make([]func(), 0)
}
