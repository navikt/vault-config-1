package vault

type Kubernetes struct {
	GenericAuth `hcl:",squash"`
}

func (k Kubernetes) GetType() string {
	return "kubernetes"
}

func (k Kubernetes) WriteStuff(c *VCClient) error {
	return nil
}
