package matter

import "context"

type SMS interface {
	SendOTP(context.Context, string, string) error
	VerifyOTP(context.Context, string, string) (bool, error)
}
type Intake struct{ MatterID, Phone, Code, SignedDocumentID string }
type Result struct {
	Authenticated          bool
	DownloadPath, FollowUp string
}

func Complete(ctx context.Context, api SMS, in Intake) (Result, error) {
	if err := api.SendOTP(ctx, in.Phone, in.MatterID); err != nil {
		return Result{}, err
	}
	ok, err := api.VerifyOTP(ctx, in.Phone, in.Code)
	if err != nil {
		return Result{}, err
	}
	if !ok {
		return Result{Authenticated: false}, nil
	}
	return Result{Authenticated: true, DownloadPath: "/matters/" + in.MatterID + "/signed/" + in.SignedDocumentID, FollowUp: "deadline review queued for " + in.MatterID}, nil
}
