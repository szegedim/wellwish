package mesh

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Mesh containers do some heavy lifting for the entire cluster.
// Individual sack and burst containers are not aware of the cluster details.
// They have only a pointer to the cluster entry point, a https site address.

// Mesh containers listen to 7778 and communicate through in EnglangRequest.
// It would not require https within the VPC, but we use TLS closure for now.
// - Mesh reads sack checkpoint backups.
// - Mesh knows where to find a sack and forwards requests to other nodes
// - Mesh can restore an entire cluster
// - Mesh sets up a node metal file with key for burst nodes
// - Burst nodes log in with the key in the metal file to mesh to get tasks to run.
// - Mesh can be on the same container as sacks or others running static code
// - Burst is running dynamic code, it exits every time after a run.

// We store checkpoints locally on each node.
// A Redis runner can pick them up and back them up regularly
// How? Potentially it is mapped to a sack and a burst with Redis client can pick it up.

// How often?
// Checkpoints too rare may lose important recent changes, ergo support costs.
// Checkpoints too frequent may require differential storage, ergo support costs.
// Differentials also tend to restore slower being eventually a downtime extender, ergo support costs.
//
// Solution: we are safe to run checkpoints as often as their collection timespan.
// This also allows consistency and hardware error checks and fixes.

// This also means that mesh is 100% letter A = Available in the CAP theorem.
// Consistency is implied by running personal cloud items independently by apikey.
// The application layer can add consistency features. We are eventually consistent.
// Partition tolerance can be implemented at the application level buying two sacks.
// The temporary nature of sacks also helps to down prioritize partition tolerance.

func Setup() {

	http.HandleFunc("/mesh.html", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		if drawing.ResetSession(w, r) != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "mesh")
	})
	http.HandleFunc("/mesh.png", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		drawing.ServeRemoteFrame(w, r, declareForm)
	})

	//http.HandleFunc("/node", func(w http.ResponseWriter, r *http.Request) {
	//	// Load and Propagate server names from api
	//	adminKey, err := management.EnsureAdministrator(w, r)
	//	management.QuantumGradeAuthorization()
	//	if err != nil {
	//		w.WriteHeader(http.StatusUnauthorized)
	//		return
	//	}
	//	address := string(drawing.NoErrorBytes(io.ReadAll(r.Body)))
	//	if address == "" {
	//		w.WriteHeader(http.StatusNoContent)
	//		return
	//	}
	//
	//	// We circle back
	//	ForwardRoundRobinRingRequest(r)
	//
	//	if r.Method == "PUT" {
	//		if Nodes[address] != "" {
	//			// No reflection, avoid hangs
	//			return
	//		}
	//		Nodes[address] = address
	//		for node, status := range Nodes {
	//			if status != "This node got an eviction notice." {
	//				go func() {
	//					NewRoundRobinCall(fmt.Sprintf("%s/node?apikey=%s", node, adminKey), "PUT", strings.NewReader(address))
	//				}()
	//			}
	//		}
	//
	//		// TODO retry propagation, if missing nodes are found
	//		// Do not retry
	//		// Retries usually just map malware errors as a unit test
	//		// Make sure that your metal is steel.
	//		//
	//	}
	//	if r.Method == "DELETE" {
	//		if Nodes[address] == "" {
	//			w.WriteHeader(http.StatusNotFound)
	//			return
	//		}
	//		if Nodes[address] == "This node got an eviction notice." {
	//			return
	//		}
	//		Nodes[address] = "This node got an eviction notice."
	//	}
	//	// There is intentionally no way to get the list of nodes. Parse checkpoint traces for debugging.
	//})

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		// Load and Propagate server names from api
		_, err := management.EnsureAdministrator(w, r)
		management.QuantumGradeAuthorization()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method == "PUT" {
			// Store locally
			UpdateIndex(r.Body)

			// Merge with existing and forward
			merged := FilterIndexEntries()

			// Propagate remotely
			ForwardRoundRobinRingRequestUpdated(r, merged)
		}
		if r.Method == "GET" {
			buf := FilterIndexEntries()
			_, _ = io.Copy(w, buf)
		}
	})

	checkpointingSetup()
}

func InitializeNodeList() {
	if len(Nodes) > 0 {
		return
	}
	nodes := map[string]string{}
	NodePattern = metadata.NodePattern
	if NodePattern != "" {
		actual := []string{NodePattern}
		for {
			next := make([]string, 0)
			for _, x := range actual {
				if strings.Contains(x, "*") {
					for i := 0; i < 10; i++ {
						next = append(next, strings.Replace(x, "*", englang.DecimalString(int64(i)), 1))
					}
				}
			}
			if len(next) == 0 {
				break
			}
			actual = next
		}
		for _, node := range actual {
			nodes[node] = "Node candidate generated by metadata node pattern."
		}
	}
	Nodes = nodes
}

func Proxy(w http.ResponseWriter, r *http.Request) error {
	apiKey := r.Header.Get("apikey")
	if apiKey == "" {
		w.WriteHeader(http.StatusNotFound)
		return fmt.Errorf("not found")
	}
	server := findServerOfApiKey(apiKey)
	if server == "" {
		w.WriteHeader(http.StatusNotFound)
		return fmt.Errorf("not found")
	}
	if englang.Synonym(Nodes[server], "This node got an eviction notice.") {
		w.WriteHeader(http.StatusGone)
		return fmt.Errorf("not found")
	}
	if strings.HasPrefix(metadata.SiteUrl, "http://") &&
		!strings.HasPrefix(server, "http://") {
		server = "http://" + server
	} else if strings.HasPrefix(metadata.SiteUrl, "https://") &&
		!strings.HasPrefix(server, "https://") {
		server = "https://" + server
	}
	original := r.URL.String()
	modified := strings.Replace(original, metadata.SiteUrl, server, 1)
	if modified == original {
		w.WriteHeader(http.StatusNotFound)
		return fmt.Errorf("not found")
	}
	b, _ := management.HttpProxyRequest(modified, r.Method, r.Body)
	// TODO Is it okay to assume a complete write with HTTP writer?
	_, _ = w.Write(b)
	return nil
}

func declareForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./billing/res/mesh.png")

		var Text = 0

		instruction := fmt.Sprintf("Set up mesh network using health check results from the nodes: %s", metadata.NodePattern)
		drawing.DeclareTextField(session, Text, drawing.ActiveContent{Text: "�" + instruction, Lines: 20, Editable: true, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 1})

	}
}
