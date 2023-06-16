package burst

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/stateful"
	"io"
	"net/http"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Burst runners run containerized applications in pull mode.

// The main design behind burst runners is that they are designed to be scalable.
// Data locality means that data is co-located with burst containers.
// Data locality is important in some cases, especially UI driven code like ours.

// Bursts are designed to handle the longer running processes with chaining.
// They do not wait, but they register pass on the results to another burst saving serverless time.

// UI applets should be low latency using just bags and direct code.
// Bursts should scale out. They are okay to be located elsewhere than the data bags.
// The reason is that large computation will require streaming, and
// streaming is driven by pipelined steps without replies and feedbacks.
// Streaming bandwidth is not affected by co-location of data and code.
// Example: 1million 100ms reads followed by 100ms compute will last 200000 seconds.
// Example: 1million 100ms reads streamed into 100ms compute will last 100000 seconds,
// even if there is an extra network latency of 150ms per burst.

// Box is a container code that waits for a single burst and exits
// The host runs an /idle?apikey=<ACTIVATION KEY> call to get a unique key
// TODO Consider using the management key here?
// The host then runs a box such as
// docker run -d --rm --restart=always --name box1 -e BURSTKEY=<key> wellwish
// It will do a curl -X GET /idle?apikey=<BURSTKEY> to fetch the instructions
// It will return the results to curl -X PUT /ilde?apikey=<BURSTKEY> after running

// There are two ways to input and output data to and from such boxes
// One is the burst `/run` request body and return.
// Keep this small as it transfers through multiple http requests.
// The other way is to pass a bag url or cloud bucket url where the box streams any input or results.

// We do not log runtime or errors, the server takes care of that.
// It is so simple that once it works it will work forever.
// The design is that it runs for 1-10 seconds.
// This is the bandwidth of a single 1 vcpu+1 gigabyte container streamed entirely.
// It is similar to serverless lambdas, but it is a bit better.
// Lambdas can wait but bursts typically use cpu bursts, and they continue in another burst.
// This makes bursts less expensive to the cloud provider using the cpu as much as possible.
// Also, bursts are not called, but they run with standard input and standard output.
// They are not an api endpoint, the api gateway is the wellwish server.
// This makes burst more secure and easier to use just like a bash script or cgi script.
// Bursts are typically docker containers with php/java/node preloaded by the taste of the cloud office cluster.
// They keep checking the frontend for new tasks, and they restart when done cleaning all interim results.

// TODO add timeout logic on paid vouchers

var startTime = time.Now()
var code = make(chan chan string)
var firstRun = true

func Setup() {
	stateful.RegisterModuleForBackup(&BurstSession)

	http.HandleFunc("/run", func(writer http.ResponseWriter, request *http.Request) {
		apiKey := request.URL.Query().Get("apikey")
		_, call := BurstSession[apiKey]
		if !call {
			management.QuantumGradeAuthorization()
			writer.WriteHeader(http.StatusPaymentRequired)
			drawing.NoErrorWrite(writer.Write([]byte("Payment required with a PUT to /run.coin")))
			return
		}

		input := drawing.NoErrorString(io.ReadAll(request.Body))
		callChannel := make(chan string)

		select {
		case <-time.After(MaxBurstRuntime):
			break
		case code <- callChannel:
			break
		}

		select {
		case <-time.After(MaxBurstRuntime):
			break
		case callChannel <- input:
			break
		}

		select {
		case <-time.After(MaxBurstRuntime + MaxBurstRuntime):
			break
		case output := <-callChannel:
			drawing.NoErrorWrite64(io.Copy(writer, bytes.NewBuffer([]byte(output))))
			break
		}
	})
	http.HandleFunc("/idle", func(writer http.ResponseWriter, request *http.Request) {
		apiKey := request.URL.Query().Get("apikey")
		if request.Method == "GET" {
			if apiKey == metadata.ActivationKey {
				// We may live without activation key
				// but this allows restricting the office cluster endpoint
				// to internal 127.0.0.1 addresses that was easier with udp.
				lock.Lock()
				idle := drawing.GenerateUniqueKey()
				ContainerRunning[apiKey] = fmt.Sprintf("Burst box %s registered at %s second.", idle, englang.DecimalString(int64(time.Now().Sub(startTime).Seconds())))
				ret := bytes.NewBufferString(idle)
				drawing.NoErrorWrite64(io.Copy(writer, ret))
				lock.Unlock()
				go func(key string) {
					if !firstRun {
						time.Sleep(MaxBurstRuntime * 2)
					}
					lock.Lock()
					ContainerRunning[key] = fmt.Sprintf("Burst box %s registered at %s second is ready.", idle, englang.DecimalString(int64(time.Now().Sub(startTime).Seconds())))
					lock.Unlock()
				}(idle)
				return
			}
			lock.Lock()
			v, ok := ContainerRunning[apiKey]
			lock.Unlock()
			if ok {
				var key, started string
				if nil == englang.Scanf1(v, "Burst box %s registered at %s second is ready.", &key, &started) {
					firstRun = false
					select {
					case <-time.After(MaxBurstRuntime):
						break
					case callChannel := <-code:
						request := <-callChannel
						lock.Lock()
						delete(ContainerRunning, apiKey)
						ContainerResults[apiKey] = callChannel
						lock.Unlock()
						go func(key string) {
							time.Sleep(MaxBurstRuntime * 2)
							lock.Lock()
							delete(ContainerResults, key)
							lock.Unlock()
						}(apiKey)
						ret := bytes.NewBufferString(request)
						drawing.NoErrorWrite64(io.Copy(writer, ret))
						break
					}
				}
			} else {
				// Not ready
			}
			return
		}
		if request.Method == "PUT" {
			// TODO Get container result
			result := drawing.NoErrorString(io.ReadAll(request.Body))
			lock.Lock()
			replyCh, ok := ContainerResults[apiKey]
			lock.Unlock()
			if ok {
				select {
				case <-time.After(10 * time.Millisecond):
					break
				case replyCh <- result:
					break
				}
				go func() {
					lock.Lock()
					delete(ContainerResults, apiKey)
					lock.Unlock()
				}()
			}
			return
		}
	})
	http.HandleFunc("/run.coin", func(w http.ResponseWriter, r *http.Request) {
		// Setup burst sessions, a range of time, when a coin can be used for bursts.
		if r.Method == "PUT" {
			coinToUse := billing.ValidatedCoinContent(w, r)
			if coinToUse != "" {
				func() {
					lock.Lock()
					defer lock.Unlock()
					// TODO generate new?
					burst := coinToUse
					// TODO cleanup
					BurstSession[burst] = englang.Printf(fmt.Sprintf("Burst chain api created from %s is %s/run.coin?apikey=%s. Chain is valid until %s.", coinToUse, metadata.Http11Port, burst, time.Now().Add(24*time.Hour).String()))
					mesh.SetExpiry(burst, ValidPeriod)
					mesh.RegisterIndex(burst)
					// TODO cleanup
					// mesh.SetIndex(burst, mesh.WhoAmI)
					management.QuantumGradeAuthorization()
					_, _ = w.Write([]byte(burst))
				}()
				return
			}
			management.QuantumGradeAuthorization()
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		if r.Method == "GET" {
			apiKey := r.URL.Query().Get("apikey")
			session, sessionValid := BurstSession[apiKey]
			if !sessionValid {
				management.QuantumGradeAuthorization()
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			management.QuantumGradeAuthorization()
			_, _ = w.Write([]byte(session))
			return
		}
	})
}
