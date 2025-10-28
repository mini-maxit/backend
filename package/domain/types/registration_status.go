package types

type RegistrationRequestStatus string

const (
	RegistrationRequestStatusApproved RegistrationRequestStatus = "approved"
	RegistrationRequestStatusRejected RegistrationRequestStatus = "rejected"
	RegistrationRequestStatusPending  RegistrationRequestStatus = "pending"
)

func ParseRegistrationRequestStatus(status string) (RegistrationRequestStatus, bool) {
	switch status {
	case string(RegistrationRequestStatusApproved):
		return RegistrationRequestStatusApproved, true
	case string(RegistrationRequestStatusRejected):
		return RegistrationRequestStatusRejected, true
	case string(RegistrationRequestStatusPending):
		return RegistrationRequestStatusPending, true
	default:
		return "", false
	}
}
