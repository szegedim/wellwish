package billing

import (
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"net/http"
	"time"
)

func IsApiKeyValid(w http.ResponseWriter, r *http.Request, validated *map[string]string, fallback func(w http.ResponseWriter, r *http.Request) error) (string, error) {
	apiKey := r.URL.Query().Get("apikey")

	if apiKey == "" {
		w.WriteHeader(http.StatusPaymentRequired)
		return "", fmt.Errorf("no apikey")
	}
	// TODO management.QuantumGradeAuthorization()
	// TODO The option is to delete/recreate a lost or stolen apikey
	// We can add another option to mask it with a newly generated one here.
	content := (*validated)[apiKey]
	if content == "" {
		err := fallback(w, r)
		if err == nil {
			return "", fmt.Errorf("handled my mesh")
		}
		w.WriteHeader(http.StatusNotFound)
		return "", fmt.Errorf("no payment")
	}
	expiry := ""
	err := englang.Scanf(content, TicketExpiry, &expiry)
	if err != nil {
		w.WriteHeader(http.StatusPaymentRequired)
		return "", fmt.Errorf("expired apikey")
	}
	expired, err := time.Parse("Jan 2, 2006", expiry)
	if err != nil {
		w.WriteHeader(http.StatusPaymentRequired)
		return "", fmt.Errorf("expiry misformatted apikey")
	}
	if time.Now().After(expired) {
		w.WriteHeader(http.StatusPaymentRequired)
		return "", fmt.Errorf("expired apikey")
	}
	return apiKey, nil
}
