## gravedigger version 0.000069
![gravedigger](gravedigger.jpg) 

waybackurls + status probe, digs for archived URLs.

 Thanks [Tomnomnom](https://github.com/Tomnomnom) for the great tool and idea

```
Usage:  gravedigger -urls domain.tld to dig only for Urls
        gravedigger -subdomains domain.tld to print out subdomains
        gravedigger -status domain.tld to perform a HTTP status check on the URLS (WARNING:Can take a long ass time)

```

# build

go build -o gravedigger -ldflags "-w -s" main.go
