package vault

import (
	"fmt"
)

type Ldap struct {
	GenericAuth     `hcl:",squash"`
	Users  []Entity `hcl:"User"`
	Groups []Entity `hcl:"group"`
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

func (l Ldap) WriteStuff(c *VCClient) error {
	if err := l.WriteUsers(c); err != nil {
		return err
	}

	if err := l.WriteGroups(c); err != nil {
		return err
	}

	return nil

}
