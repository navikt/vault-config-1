package vault

import (
	"fmt"
)

type Github struct {
	GenericAuth `hcl:",squash"`
	Users []struct {
		Name    string                 `hcl:",key"`
		Options map[string]interface{} `hcl:"options"`
	} `hcl:"users,ommitempty"`
	Groups []struct {
		Name    string                 `hcl:"name"`
		Options map[string]interface{} `hcl:"options"`
	} `hcl:"teams,ommitempty"`
}

func (g Github) WriteUsers(c *VCClient) error {
	userPath := fmt.Sprintf("%s/map/users", Path(g))

	for _, v := range g.Users {
		path := fmt.Sprintf("%s/%s", userPath, v.Name)
		_, err := c.Logical().Write(path, v.Options)
		if err != nil {
			return fmt.Errorf("Error writing value to Vault: %v", err)
		}
	}

	return nil
}

func (g Github) WriteGroups(c *VCClient) error {
	groupPath := fmt.Sprintf("%s/map/teams", Path(g))

	for _, v := range g.Groups {
		path := fmt.Sprintf("%s/%s", groupPath, v.Name)
		_, err := c.Logical().Write(path, v.Options)
		if err != nil {
			return fmt.Errorf("Error writing value to Vault: %v", err)
		}
	}

	return nil
}
