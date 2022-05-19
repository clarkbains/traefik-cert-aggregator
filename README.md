# Traefik Cert Aggregator

## Intro

This little go program is designed to be an easily extensible solution for certificate storage when using operating multiple community traefik instances to provide HA, with letsencrypt. The traefik team has come out stating that this will not be added (read the conversation here [#5426: 2.0 is released, is cluster HA ready?](https://github.com/traefik/traefik/issues/5426))

I have yet to find any solutions that seem fully integrated, so this and the complimentary [le-exporter](github.com/clarkbains/le-exporter) are my attempt to create a solution.

This is built as a generic and modular certificate aggregator, taking certificates from any number of sources, and putting them into any other number of sinks, mostly because I wanted to practice my go skills. A vault kv2 source is integrated, and a traefik sink, which writes the certificates out to disk, and generates a yaml file, which a traefik dynamic file provider can watch. This way, as new certs are sourced, they are written, and traefik uses them.

See `./deployment` for nomad, consul, and vault job descriptions/acl policies

## Background

With traefik, you get easy letencrypt integration, though you cannot use it in HA. Lets encrypt issues challenges to servers, and expects that challenge to be repeated back for requests from other lets encrypt servers. This makes it challenging to do in traefik running in HA since the instances need to coordinate the active challenges, so any could respond. Traefik community also doesn't have easy ways to get certificates from a centralized source, so even if the letsencrypt challenges could be abstracted away, it would be difficult to get the certificates to every instance.

The other reverse proxy I tried which integrated well with the Hashicorp stack I am using (consul, vault), is [fabio](github.com/fabiolb/fabio). It is completely stateless, and can retrieve certificates from the vault k/v store. Unfortunately, it cannot automatically generate certs using letsencrypt, and lacks some of the middleware options in traefik, like forwardAuth.

As a solution, I wrote this, and [le-exporter](github.com/clarkbains/le-exporter). le-exporter is for generating certificates using letsencrypt, and uploading them into my vault instance. It can run on one, or multiple nodes in your network, however as vault manages the actual storage of the secrets, it is not a critical piece of infrastructure. To compliment that script, this program exists, to download certificates out of vault, and put them into each traefik instance.

## TODO
I started with a system to configure the individual sources and sinks, but the current configuration is not very flexible, nor is the help or documentation done for any config.
