package main

import (
	"context"
	"encoding/json"
	"legaltech-sms-otp/internal/infrai"
	"legaltech-sms-otp/internal/matter"
	"log"
	"net/http"
)

func main() {
	api, err := infrai.New()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var in matter.Intake
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			http.Error(w, "invalid request", 400)
			return
		}
		result, err := matter.Complete(context.Background(), api, in)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	log.Println("legal OTP service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
