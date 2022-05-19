# Create a vault policy for this with 'vault write consul/roles/traefik policies="traefik-traefik"'
# Call this traefik-traefik

key_prefix "traefik" {
  policy = "write"
}

service "traefik" {
  policy = "write"
}

service_prefix "" {
  policy = "read"
}

# These may not be needed
agent_prefix "" {
  policy = "read"
}

node_prefix "" {
  policy = "read"
}