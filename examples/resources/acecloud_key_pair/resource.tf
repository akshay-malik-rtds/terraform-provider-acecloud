# Generate a new key pair (private key returned in state)
resource "acecloud_key_pair" "generated" {
  name = "deploy-key"
}

# Or import an existing public key
resource "acecloud_key_pair" "imported" {
  name       = "my-laptop"
  public_key = file("~/.ssh/id_ed25519.pub")
}
