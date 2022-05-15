# Traefik Cert Aggregator

This little go program is designed to be an easily extensible solution for certificate storage when using operating multiple community traefik instances to provide HA. The letsencrypt team has come out stating that this will not be added (read the conversation here [#5426: 2.0 is released, is cluster HA ready?](https://github.com/traefik/traefik/issues/5426))

I have yet to find any solutions that seem fully integrated, so this and the complimentary [le-exporter]() (not published yet) are my attempt to create a solution.

With traefik, you get easy letencrypt integration, though you cannot use it in HA. Lets encrypt issues challenges to servers, and expects that challenge to be repeated back for requests from other lets encrypt servers. This makes it challenging to do in traefik running in HA since the instances need to coordinate the active challenges, so any could respond. Traefik community also doesn't have easy ways to get certificates from a centralized source, so even if the letsencrypt challenges could be abstracted away, it would be difficult to get the certificates to every instance.

The other reverse proxy I tried which integrated well with the Hashicorp stack I am using (consul, vault), is [fabio](github.com/fabiolb/fabio). It is completely stateless, and can retrieve certificates from the vault k/v store. Unfortunately, it cannot automatically generate certs using letsencrypt, and lacks some of the middleware options in traefik, like forwardAuth.

As a solution, I wrote this, and le-exporter. le-exporter is for generating certificates using letsencrypt, and uploading them into my vault instance. It can run on one, or multiple nodes in your network, however as vault manages the actual storage of the secrets, it is not a critical piece of infrastructure. To compliment that script, this program exists, to download certificates out of vault, and put them into each traefik instance.

This is only a prototype, but the idea would be to write the files out to disk in a volume shared with the traefik docker container, and have it generate a yaml file for traefik describing where all the certificates are. Traefik should then be able to pick this up, and start using the new certs. Since this is mostly just fetching data, it can be run alondside each traefik instance, with no concurrency concerns.