# DRE

Docker Registry Exploration

## Concept

It connects to docker registries to queries all availables repos and tags.
The tool then proceeds to download each files of **the last 2 layers** for each tags.
It also stores the image config in config.json for further analysis.

## Usage

```sh
$ ./dre "https://myregistry.company.com" ./explore [image-name] [tag-name]
$ find ./explore|grep json
test/xxxx/Debug-22/root/app/appsettings.SIT-2.json
test/xxxx/Debug-22/root/app/appsettings.json
test/xxxx/Debug-22/root/app/ApolloDeliveryBookingAvailableWatcher.deps.json
test/xxxx/Debug-22/root/app/appsettings.LIVE-1.json
```

Will explore `https://myregistry.company.com` and store the results in `./explore`
