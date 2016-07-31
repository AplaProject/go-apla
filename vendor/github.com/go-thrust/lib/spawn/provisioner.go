package spawn

type Provisioner interface {
	Provision() error
}

// Package global provisioner, use SetProvisioner to set this variable
var provisioner Provisioner

// Sets the current provision to use during spawn.Run()
func SetProvisioner(p Provisioner) {
	provisioner = p
}

/*
Default implementation of Provisioner
*/
type ThrustProvisioner struct{}

func NewThrustProvisioner() ThrustProvisioner {
	return ThrustProvisioner{}
}

func (tp ThrustProvisioner) Provision() error {
	return Bootstrap()
}
