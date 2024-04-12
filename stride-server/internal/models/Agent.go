package models

type Agent struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	IP       string   `json:"ip"`
	Created  string   `json:"created"`
	Tags     []string `json:"tags"`
	Software []string `json:"software"`
	SSHKeys  []string `json:"ssh_keys"` // Update to a slice of strings to store multiple SSH key IDs
	Size     string   `json:"size"`     // Add a new field to store the droplet size
}

// SSHKey represents an SSH key in your system, mirroring the structure returned by DigitalOcean.
type SSHKey struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    PublicKey  string `json:"public_key,omitempty"` // Include if you plan to use/display the public key.
}