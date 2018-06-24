package vault

import (
	"fmt"
)

type Github struct {
	GenericAuth    `hcl:",squash"`
	Users []Entity `hcl:"users,ommitempty"`
	Teams []Entity `hcl:"teams,ommitempty"`
}

func (g Github) GetType() string {
	return "github"
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

func (g Github) WriteTeams(c *VCClient) error {
	groupPath := fmt.Sprintf("%s/map/teams", Path(g))

	for _, v := range g.Teams {
		path := fmt.Sprintf("%s/%s", groupPath, v.Name)
		_, err := c.Logical().Write(path, v.Options)
		if err != nil {
			return fmt.Errorf("Error writing value to Vault: %v", err)
		}
	}

	return nil
}

func (l Github) WriteStuff(c *VCClient) error {
	if err := l.WriteUsers(c); err != nil {
		return err
	}

	if err := l.WriteTeams(c); err != nil {
		return err
	}
	return nil

}
