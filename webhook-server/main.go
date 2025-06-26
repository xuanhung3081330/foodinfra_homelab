package main

import (
	"encoding/json" // used to decode/encode json object
	"fmt"           // used to print something
	"net/http"      // used to handle endpoint for api

	admissionv1 "k8s.io/api/admission/v1"         // the same as api group in k8s
	corev1 "k8s.io/api/core/v1"                   // the same as api group in k8s
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" // the same as api group in k8s
)

func main() {
	http.HandleFunc("/validate", handleValidate)
	fmt.Println("Starting server on port 8443...")
	http.ListenAndServeTLS(":8443", "/tls/tls.crt", "/tls/tls.key", nil)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	// http.ResponseWriter: an interface and interfaces in Go are already references under the hood.

	var review admissionv1.AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Could not decode admission review", http.StatusBadRequest)
		return
	}

	var pod corev1.Pod
	if err := json.Unmarshal(review.Request.Object.Raw, &pod); err != nil {
		http.Error(w, "Could not parse pod object", http.StatusInternalServerError)
		return
	}

	allowed := true
	message := "Pod accepted"
	for _, container := range pod.Spec.Containers {
		if container.Image == "" || len(container.Image) >= 7 && container.Image[len(container.Image)-7:] == ":latest" {
			allowed = false
			message = fmt.Sprintf("Image '%s' uses ':latest', which is not allowed", container.Image)
			break
		}
	}

	response := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     review.Request.UID,
			Allowed: allowed,
			Result: &metav1.Status{
				Message: message,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
