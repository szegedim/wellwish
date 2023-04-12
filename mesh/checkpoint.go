package mesh

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

func runCheckPoint() string {
	var checkpoint = bytes.Buffer{}
	management.CheckpointFunc("GET", &checkpoint, nil)

	// Checkpoint
	sackId := drawing.GenerateUniqueKey()
	fileName := path.Join("/tmp", sackId)
	_ = os.WriteFile(fileName, checkpoint.Bytes(), 0700)
	_ = os.Remove("/tmp/checkpoint")
	_ = os.Link(fileName, "/tmp/checkpoint")

	// Update public facing indexes from checkpoint
	UpdateIndex(drawing.NoErrorFile(os.Open("/tmp/checkpoint")))

	// Propagate remotely
	index := FilterIndexEntries()
	url3 := management.AddAdminForUrl(fmt.Sprintf("%s/index", metadata.SiteUrl))
	NewRoundRobinCall(url3, "PUT", index)

	fmt.Printf("Health check succeeded %s checkpoint file %s ...\n", time.Now().Format("15:04:05"), path.Join("/tmp", drawing.RedactPublicKey(sackId)))
	return sackId
}

func checkpointingSetup() {
	http.HandleFunc("/site.backup", func(w http.ResponseWriter, r *http.Request) {
		// Capture a checkpoint of the state of each node on the cluster
		if r.Method == "PUT" {
			apiKey, err := management.EnsureAdministrator(w, r)
			management.QuantumGradeAuthorization()
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			for node := range Nodes {
				// Make sure your ops works
				sackId, _ := management.HttpProxyRequest(fmt.Sprintf("http://%s/node.checkpoint?apikey=%s", node, apiKey), "GET", nil)
				if sackId != nil && len(sackId) > 0 {
					Index[string(sackId)] = node
				}
			}
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/site.restore", func(w http.ResponseWriter, r *http.Request) {
		// Capture a checkpoint of the state of each node on the cluster
		if r.Method == "PUT" {
			_, err := management.EnsureAdministrator(w, r)
			management.QuantumGradeAuthorization()
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ForwardRoundRobinRingRequest(r)
			management.CheckpointFunc("PUT", nil, r.Body)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/node.checkpoint", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			// Create a new checkpoint file
			_, err := management.EnsureAdministrator(w, r)
			management.QuantumGradeAuthorization()
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sackId := runCheckPoint()
			_, _ = w.Write([]byte(fmt.Sprintf("%s/node.checkpoint?apikey=%s", metadata.SiteUrl, sackId)))
		}
		if r.Method == "GET" {
			// Get specified checkpoint file
			apiKey := r.URL.Query().Get("apikey")
			if apiKey == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			management.QuantumGradeAuthorization()

			sackId := apiKey
			fileName := path.Join("/tmp", sackId)
			temp := path.Join("/tmp", fmt.Sprintf("%s_%s", sackId, drawing.RedactPublicKey(drawing.GenerateUniqueKey())))

			_ = os.Remove(temp)
			_ = os.Link(fileName, temp)
			// No streaming so that the node always have 50% free memory for latency
			requested := drawing.NoErrorBytes(os.ReadFile(temp))
			if requested == nil || len(requested) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write(requested)
			_ = os.Remove(temp)
		}
	})

	http.HandleFunc("/node.restore", func(w http.ResponseWriter, r *http.Request) {
		// Restore stateful state to a checkpoint (sacks)
		if r.Method == "PUT" {
			_, err := management.EnsureAdministrator(w, r)
			management.QuantumGradeAuthorization()
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if len(Index) > 0 {
				w.WriteHeader(http.StatusConflict)
				return
			}

			management.CheckpointFunc("PUT", nil, r.Body)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	go func() {
		for {
			// We prioritize Availability from CAP theorem
			// Partition tolerance and consistency can be applied on the top,
			// if an available system is given.
			runCheckPoint()
			time.Sleep(metadata.CheckpointPeriod)
		}
	}()
}

func ActivateSite() {
	// Set up round-robin by adding ourselves as a node
	//_, _ = management.HttpProxyRequest(fmt.Sprintf("%s/node?apikey=%s", metadata.NodeUrl, managementKey), "PUT", bytes.NewBuffer([]byte(metadata.NodeUrl)))

	url1 := fmt.Sprintf("%s/activate?activationkey=%s", metadata.SiteUrl, management.SiteActivationKey)
	url1 = management.AddAdminForUrl(url1)
	NewRoundRobinCall(url1, "GET", &bytes.Buffer{})
}

func NewRoundRobinCall(url1 string, method string, body io.Reader) {
	session := drawing.GenerateUniqueKey()
	next, nextNext, err := roundRobinRing("", session)
	if err != nil {
		return
	}
	url := fmt.Sprintf("%s&session=%s&next=%s", strings.Replace(url1, metadata.SiteUrl, next, 1), session, nextNext)
	_, err = management.HttpProxyRequest(url, method, body)
	if err != nil {
		Rings[session] = Rings[session] + "Request failed."
	}
	Rings[session] = strings.TrimRight(Rings[session], "\n") + url + "\n"
}

func ForwardRoundRobinRingRequest(r *http.Request) {
	ForwardRoundRobinRingRequestUpdated(r, r.Body)
}

func ForwardRoundRobinRingRequestUpdated(r *http.Request, updated io.Reader) {
	next := r.URL.Query().Get("next")
	session := r.URL.Query().Get("session")
	if next == "" || session == "" {
		return
	}

	u := r.URL
	next, nextNext, err := roundRobinRing(next, session)
	if err != nil {
		//Rings[session] = strings.TrimRight(Rings[session], "\n") + r.URL.Path + "\n"
		return
	}
	u.Query().Set("session", session)
	u.Query().Set("next", nextNext)
	u.Host = next
	_, err = management.HttpProxyRequest(u.String(), r.Method, updated)
}

//func roundRobinRingRequest(r *http.Request) (string, string, string, error) {
//	next := r.URL.Query().Get("next")
//	session := r.URL.Query().Get("session")
//
//	next, nextNext, session, err := roundRobinRing(next, session)
//	return next, nextNext, session, err
//}

func roundRobinRing(next string, ringSession string) (string, string, error) {
	InitializeNodeList()
	if ringSession != "" && Rings[ringSession] != "" {
		var ring, started, base, begin, end string
		if englang.ScanfContains(Rings[ringSession], "Ring %s started on %s at %s ns", &begin, &ring, &started, &base, &end) == nil {
			elapsed := englang.DecimalString(time.Now().UnixNano() - englang.Decimal(base))
			Rings[ringSession] = fmt.Sprintf("Ring %s started on %s. It finished in %s ns %s\n", ringSession, time.Now().Format("Jan 2, 2006"), elapsed, end)
		}
		// Finished a circular propagating call
		return "", "", fmt.Errorf("finished")
	}
	nodeNames := make([]string, 0)
	nodes := Nodes
	for node := range nodes {
		nodeNames = append(nodeNames, node)
	}
	sort.Strings(nodeNames)

	nextNext := ""
	for i := 0; i < len(nodeNames); i++ {
		if nodeNames[i] == next || next == "" {
			_, err := management.HttpProxyRequest(fmt.Sprintf("%s/healthz", nodeNames[i]), "GET", nil)
			if err == nil {
				Rings[ringSession] = fmt.Sprintf("Ring %s started on %s at %s ns", ringSession, time.Now().Format("Jan 2, 2006"), englang.DecimalString(time.Now().UnixNano()))
				next = nodeNames[i]
				nextNext = nodeNames[(i+1)%len(nodeNames)]
				return next, nextNext, nil
			}
		}
	}
	// TODO no cluster?
	return "", "", fmt.Errorf("finished")
}
