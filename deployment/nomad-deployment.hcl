
job "Network-Ingress" {
  region      = "global"
  datacenters = ["cwdc"]
  constraint {    
    attribute = "${node.unique.name}"    
    value     = "rp-1"    
  }

  group "Network-Ingress" {
    count = 1
    network {
      port "http" {
        static = 80
      }
      
      port "https" {
        static = 443
      }

      port "internal-https" {
        static = 4443
      }

    }
    

    service {
      name = "traefik"
      check {
        name     = "alive"
        type     = "tcp"
        port     = "http"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "cert-puller" {
      driver = "docker"
      
      config {
        image        = "ghcr.io/clarkbains/cert-agg:latest"
        network_mode = "host"
      }

      vault {        
        policies = ["cert-aggregator"]
      }
      
      template {
        data = <<EOH
VAULT_ADDR=https://192.168.25.137:8200
TRAEFIK_BASE=/alloc/data/traefik
        EOH
        destination = "local/file.env"
        change_mode   = "restart"
        env         = true
        }

    }

    task "traefik" {
      driver = "docker"
      config {
        image        = "traefik:latest"
        network_mode = "host"
        volumes = [
          "local/traefik.toml:/etc/traefik/traefik.toml"
          ]
      }

      vault {        
        policies = ["traefik"]
      }


      template {
        data = <<EOF
# Allow backends to use self signed ssl certs (Needed for waypoint)
[serversTransport]
  insecureSkipVerify = true
[entryPoints]
    [entryPoints.http]
    address = ":80"
    [entryPoints.https]
    address = ":443"
    [entryPoints.internal-secure]
    address = ":4443"

[api]
    dashboard = true

# Enable Consul Catalog configuration backend.
[providers.consulCatalog]
    prefix           = "traefik"
    exposedByDefault = false
    [providers.consulCatalog.endpoint]
      address = "127.0.0.1:8500"
      scheme  = "http"
[providers.consul]
    rootKey = "traefik"
    endpoints = ["127.0.0.1:8500"]
    token = "{{ with secret "consul/creds/traefik"}}{{.Data.token}}{{ end }}"
[providers.file]
    filename = "/alloc/data/traefik/traefik.yaml"
EOF
        destination = "local/traefik.toml"
        change_mode   = "restart"
      }
 
      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
