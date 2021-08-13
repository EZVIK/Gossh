package service

type CMD struct {
	Namespace string `validate:"required" json:"namespace"`
	IP        string `validate:"required" json:"ip"`
	Command   string `validate:"required" json:"cmd"`
}
