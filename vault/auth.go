package vault

import (
	"fmt"
	"strings"
	"github.com/fatih/structs"
)

// AuthType defines an interface for dealing with Auth backends
type AuthType interface {
	Describe() string
	GetPath() string
	GetType() string
	getAuthConfig() map[string]interface{}
	getAuthMountConfig() map[string]interface{}
	Configure(c *VCClient) error
	TuneMount(c *VCClient, path string) error
	WriteStuff(c *VCClient) error
}

type GenericAuth struct {
	Path        string                 `hcl:"path"`
	Description string                 `hcl:"description"`
	AuthConfig  map[string]interface{} `hcl:"authconfig"`
	MountConfig struct {
		DefaultLeaseTTL string `hcl:"default_lease_ttl" mapstructure:"default_lease_ttl"`
		MaxLeaseTTL     string `hcl:"max_lease_ttl" mapstructure:"max_lease_ttl"`
	} `hcl:"mountconfig"`
}

type Entity struct {
	Name    string                 `hcl:",key"`
	Options map[string]interface{} `hcl:"options"`
}

func (g GenericAuth) GetPath() string {
	return g.Path
}

func (g GenericAuth) Describe() string {
	return g.Description
}

func (g GenericAuth) TuneMount(c *VCClient, path string) error {
	return c.TuneMount(path, structs.Map(g.MountConfig))
}

func (g GenericAuth) Configure(c *VCClient) error {
	path := fmt.Sprintf("auth/%s/config", g.GetPath())
	_, err := c.Logical().Write(path, g.AuthConfig)
	if err != nil {
		return fmt.Errorf("Error writing auth config: %v", err)
	}

	return nil
}

func (g GenericAuth) getAuthConfig() map[string]interface{} {
	return g.AuthConfig
}

func (g GenericAuth) getAuthMountConfig() map[string]interface{} {
	return ConvertMapStringInterface(g.MountConfig)
}

// AuthExist checks for the existance of an Auth mount
func (c *VCClient) AuthExist(name string) bool {
	auth, err := c.Sys().ListAuth()
	if err != nil {
		return false
	}
	for a := range auth {
		if strings.TrimSuffix(a, "/") == name {
			return true
		}
	}

	return false
}

// Path will return the path of an Auth backend
func Path(a AuthType) string {
	return fmt.Sprintf("auth/%s", a.GetPath())
}

// AuthEnable enables an auth backend
func (c *VCClient) AuthEnable(a AuthType) error {
	if err := c.Sys().EnableAuth(a.GetPath(), a.GetType(), a.Describe()); err != nil {
		return err
	}

	return nil
}

// AuthConfigure sets the configuration for an auth backend
func (c *VCClient) AuthConfigure(a AuthType) error {
	if err := a.WriteStuff(c); err != nil {
		return err
	}

	if err := a.TuneMount(c, Path(a)); err != nil {
		return err
	}

	if err := a.Configure(c); err != nil {
		return err
	}

	return nil
}

func EnableAndConfigure(a AuthType, c *VCClient) error {
	if !c.AuthExist(a.GetPath()) {
		if err := c.AuthEnable(a); err != nil {
			return fmt.Errorf("Error enabling auth mount: %v", err)
		}
	}
	if err := c.AuthConfigure(a); err != nil {
		return fmt.Errorf("Error configuring auth mount: %v", err)
	}

	return nil
}
