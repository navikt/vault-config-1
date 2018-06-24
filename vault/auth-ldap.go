package vault

import (
	"fmt"

)

type Ldap struct {
	GenericAuth `hcl:",squash"`
	Users []struct {
		Name    string                 `hcl:",key"`
		Options map[string]interface{} `hcl:"options"`
	} `hcl:"User"`
	Groups []struct {
		Name    string                 `hcl:",key"`
		Options map[string]interface{} `hcl:"options"`
	} `hcl:"group"`
}

func (l Ldap) WriteUsers(c *VCClient) error {
	userPath := fmt.Sprintf("%s/users", Path(l))
	for _, v := range l.Users {
		if v.Name != "" {
			path := fmt.Sprintf("%s/%s", userPath, v.Name)
			_, err := c.Logical().Write(path, v.Options)
			if err != nil {
				return fmt.Errorf("Error writing value to Vault: %v", err)
			}
		}
	}

	return nil
}

func (l Ldap) WriteGroups(c *VCClient) error {
	groupPath := fmt.Sprintf("%s/groups", Path(l))

	for _, v := range l.Groups {
		path := fmt.Sprintf("%s/%s", groupPath, v.Name)
		_, err := c.Logical().Write(path, v.Options)
		if err != nil {
			return fmt.Errorf("Error writing value to Vault: %v", err)
		}
	}

	return nil
}
