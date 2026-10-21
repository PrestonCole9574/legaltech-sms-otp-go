package matter

import (
	"context"
	"testing"
)

type fake struct {
	verified bool
	sent     string
}

func (f *fake) SendOTP(_ context.Context, phone, _ string) error       { f.sent = phone; return nil }
func (f *fake) VerifyOTP(_ context.Context, _, _ string) (bool, error) { return f.verified, nil }
func TestCompleteRequiresVerifiedCode(t *testing.T) {
	for _, tc := range []struct {
		name     string
		verified bool
		auth     bool
	}{{"accepted", true, true}, {"rejected", false, false}} {
		t.Run(tc.name, func(t *testing.T) {
			api := &fake{verified: tc.verified}
			got, err := Complete(context.Background(), api, Intake{MatterID: "M-17", Phone: "+15550001111", Code: "246810", SignedDocumentID: "doc-9"})
			if err != nil || got.Authenticated != tc.auth {
				t.Fatalf("got %+v err %v", got, err)
			}
			if api.sent == "" {
				t.Fatal("otp was not sent")
			}
		})
	}
}
