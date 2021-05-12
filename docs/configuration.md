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
        blacklist http://url-to-my-blocklists.com/list1.txt
        blacklist http://url-to-my-blocklists.com/list2.txt
        default-lists
        block google.com
        permit googleadservices.com
        target 10.133.7.8
        target-ipv6 ::1
   }
   # Other directives have been omitted
}
```

#### Configuration options

First of all: A URL has to be of the schemes `http` and `https` for accessing remote files using HTTP and `file` to load local files. I.e. lists on the local filesystem.

Some Examples:
- Http: `http://mydomain.com/blacklist.txt`
- Https: `https://secure.mydomain.com/blacklist.txt`
- File: `file:///home/chris/blacklist.txt`

- `blacklist <LIST URL>` Add a URL of a file to load Blacklist entries from
- `whitelist <LIST URL>` Add a URL of a file to load whitelist entries from
- `default-lists` Readds the default hostlists to the internal list of blocklists.
    - This command is needed if you want to add custom blocklists and you want to also use the default ones.
    - To see a List of the Blacklist URLs click [here](lists.md)
- `strict-default-lists` Use the strict default blocklists, instead of the more soft ones.
    - Also adds a default whitelist, which can be found in this repository (`/lists/strict-whitelist.txt`) to prevent blocking of popular domains such as Facebook or Amazon.
    - To see a List of the Blacklist URLs click [here](lists.md)
- `unfiltered-strict-default-lists` just like `strict-default-lists` but here the afforementioned default whitelist is not added
- `target <IPv4 IP>` defines the target ip to which blocked domains should resolve to if a A record is requested
- `target-ipv6 <IPv6 IP>` defines the target IPv6 address to which blocked domains should resolve to if a AAAA record is requested
- `disable-auto-update` Turns off the automatic update of the blocklists every 24h (can be changed)
- `log` Print a message every time a request gets blocked
- `auto-update-interval <INTERVAL>` Allows the modification of the interval between http blocklist updates
    - This operation uses Golangs `time.ParseDuration()` function in order to parse the duration.
    Please ensure the specified duration can be parsed by this operation. Please refer to [here](https://golang.org/pkg/time/#ParseDuration).
    - This gets ignored if the automatic blocklist updates have been disabled
    - The default value is 24 hours
- `file-auto-update-interval` Just like `auto-update-interval` just for lists in the local file system
    - The default value is 1 minute
- `list-store <FILEPATH FOR PERSISTED LISTS>` This option enables persisting of the HTTP lists
  to prevent a automatic redownload everytime CoreDNS restarts. The lists get persisted everytime a update get performed.
    - If autoupdates have been turned off the list will be reloaded every time the application launches.
    Making this option pretty useless for this kind of configuration.
- `permit <QNAME>` and `block <QNAME>` Allows the explicit whitelisting or blacklisting of specific qnames. If a qname is on the whitelist it will not be blocked. 
- `permit-regex <REGEX>` and `block-regex <REGEX>` identical to the regular whitelist and blacklist options. But instead of blocking a specific qname blocking is done for a regular expression. Yo might want to define exceptions to a regex blacklist entry. This can be done by using eitehr the `whitelist` or `whitelist-regex` options. 
