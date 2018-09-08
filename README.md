# CoreDNS - ads Plugin

DNS AdBlocker plugin for CoreDNS.

## Compiling

First get the CoreDNS source code by running
```bash
go get github.com/coredns/coredns
```



```bash
sed -i 's|hosts:hosts|ads:github.com/c-mueller/ads\nhosts:hosts|g' plugin.cfg
```

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

You can see the supported configuration option in the following Corefile

```
.:53 {
   ads {
        list http://url-to-my-blocklists.com/list1.txt # Define custom Blocklists
        list http://url-to-my-blocklists.com/list2.txt
        default-lists # Use the default blocklists
        target 10.133.7.8 # The IPv4 Address (IPv6 is currently unsupported) to point blocked requests at
   }
   # Other directives have been omitted
}
```

A note when setting custom blocklists:
If the configuration contains one or more `list` options, the plugin will not load the default blocklists anymore.
In order to still keep those add the `default-lists` option.
If no list is added the default lists will be used. It is not necessary to explicitly set them.