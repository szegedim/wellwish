package mesh

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net"
	"net/http"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Mesh module functions do some heavy lifting for the entire cluster.
// Individual sack and burst containers are not aware of the cluster details.
// They have only a pointer to the cluster entry point, a https site address.

// Mesh containers listen to 7777 and communicate through Englang.
// Most cloud providers do not require https within the VPC. //TODO Is this still the case?
// - Mesh reads sack checkpoint backups.
// - Mesh knows where to find a sack and forwards requests to other nodes using Index.
// - Mesh can restore an entire cluster from and Englang backup file.
// - Mesh sets up a node metal file with keys for burst nodes.
// - Burst nodes log in with the key in the metal file to get tasks to run.
// - Mesh runs within the stateful containers of each node.
// - Burst is running dynamic code, it restarts every time after a run making a clean state.

// We store checkpoints locally on each node periodically.
// The period is set in the metadata.
// A Redis runner can pick them up using a simple sack GET and back them up regularly.
// How? It is mapped to a sack and a burst with Redis client can pick it up.

// How often?
// Checkpoints too rare may lose important recent changes, ergo support costs.
// Checkpoints too frequent may require differential storage, ergo support costs.
// Differentials also tend to restore slower being eventually a downtime extender, ergo support costs.

// Solution: we are safe to run checkpoints as often as the time to collect them allows.
// This also allows consistency and hardware error checks and fixes.

// This also means that mesh is 100% letter A = Available in the CAP theorem.
// Consistency is implied by running personal cloud items independently by apikey.
// The application layer can add consistency features. We are eventually consistent.
// Partition tolerance can be implemented at the application level buying two sacks.
// The temporary nature of sacks also helps to down prioritize partition tolerance.

// We use just a node pattern instead of having configuration to add each node.
// This allows simple node addition and removal.
// Adding a node is as simple as turning it on with the activation key propagated from the existing cluster.
// Removing a node is simple. Mark it as "This node got an eviction notice."
// TODO It is easier to add port 7778 for stateful writes and disable it in the load balancer.
// TODO It is easier to disable sack PUT requests i.e. /tmp in the load balancer or firewall.
// It can be turned off at the standard expiry time, when stateful sacks, etc. expired.

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
	actual := []string{NodePattern}
	if strings.Contains(NodePattern, "*") {
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
	}
	if strings.Contains(NodePattern, "/") && !strings.Contains(NodePattern, "//") {
		actual = getCidrAddresses(NodePattern)
	}
	for _, node := range actual {
		if node != "" {
			nodes[node] = "Node candidate generated by metadata node pattern."
		}
	}

	Nodes = nodes
}

func getCidrAddresses(cidr string) []string {
	nextIp := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	_, ipNet, _ := net.ParseCIDR(cidr)
	ret := make([]string, 0)
	ip := ipNet.IP
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); nextIp(ip) {
		ret = append(ret, ip.String())
	}
	return ret
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
		drawing.PutText(session, Text, drawing.Content{Text: "�" + instruction, Lines: 20, Editable: true, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 1})

	}
}
