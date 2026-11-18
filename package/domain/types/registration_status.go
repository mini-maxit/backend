package types

type RegistrationRequestStatus string

const (
	RegistrationRequestStatusPending  RegistrationRequestStatus = "pending"
	RegistrationRequestStatusRejected RegistrationRequestStatus = "rejected"
)

// RegistrationStatus is the current user's standing towards a public contest.
type RegistrationStatus string

const (
	RegistrationStatusRegistered         RegistrationStatus = "registered"
	RegistrationStatusCanRegister        RegistrationStatus = "canRegister"
	RegistrationStatusAwaitingApproval   RegistrationStatus = "awaitingApproval"
	RegistrationStatusRejected           RegistrationStatus = "rejected"
	RegistrationStatusRegistrationClosed RegistrationStatus = "registrationClosed"
)
