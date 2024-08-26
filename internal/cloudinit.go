package internal

type CloudInitConfiguration interface {
	Build(configuration Configuration) error
	Validate() error
}

// Validate audits the whole freyja configuration for content mistakes
func (c CloudInitData) Validate() error {
	return nil
}

// Build generate the configuration from a file
func (c CloudInitData) Build(configuration Configuration) error {
	c.setDefaultValues()
	return nil
}

// setDefaultValues set values to parameters that have not been configured but are still required
// for libvirt
func (c CloudInitData) setDefaultValues() {
	c.Output.All = ">> /var/log/cloud-init.log"
}
