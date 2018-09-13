# CoreDNS - ads Plugin

DNS AdBlocker plugin for CoreDNS.

## Compiling

First get the CoreDNS source code by running, after you cloned this repository into the proper path in your `GOPATH`
```bash
go get github.com/coredns/coredns
```

Then navigate to the coredns directory
```bash
cd $(go env GOPATH)/src/github.com/coredns/coredns
```

Next update
 the `plugin.cfg` in the root of the coredns repository

```bash
sed -i 's|hosts:hosts|ads:github.com/c-mueller/ads\nhosts:hosts|g' plugin.cfg
```

Finally run `make` to build CoreDNS with the `ads` plugin

## Configuring

### Default settings

Running the `ads` plugin with all defaults is done by just adding the `ads` keyword to your Corefile.

For example:
```
.:53 {
    ads
    forward . 1.1.1.1
    log
    errors
}
```

### Configuring the `ads` plugin

You can see a more complex `ads` configuration in the following Corefile

```
.:53 {
   ads {
        list http://url-to-my-blocklists.com/list1.txt
        list http://url-to-my-blocklists.com/list2.txt
        default-lists
        target 10.133.7.8
   }
   # Other directives have been omitted
}
```

#### Configuration options

- `list <LIST URL>` HTTP(S)-URL to a hostlist to Block
- `default-lists` Readds the default hostlists to the internal list of blocklists.
    - This command is needed if you want to add custom blocklists and you want to also use the default ones
- `target <IPv4 IP>` defines the target ip to which blocked domains should resolve to
- `disable-auto-update` Turns off the automatic update of the blocklists every 24h (can be changed)
- `log` Print a message every time a request gets blocked
- `auto-update-interval <INTERVAL>` Allows the modification of the interval between blocklist updates
    - This operation uses Golangs `time.ParseDuration()` function in order to parse the duration.
    Please ensure the specified duration can be parsed by this operation. Please refer to [here](https://golang.org/pkg/time/#ParseDuration).
    - This gets ignored if the automatic blocklist updates have been disabled
- `blocklist-path <FILEPATH FOR PERSISTED BLOCKLIST>` This option enables persisting of the blocklist
  to prevent a automatic redownload everytime CoreDNS restarts. The lists get persisted everytime a update get performed.
    - If autoupdates have been turned off the list will be reloaded every time the application launches.
    Making this option pretty useless for this kind of configuration.