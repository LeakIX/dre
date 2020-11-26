# DRE

Docker Registry Exploration

## Usage

```sh
$ ./dre "https://myregistry.company.com" ./explore
$ find ./explore|grep json
test/xxxx/Debug-22/root/app/appsettings.SIT-2.json
test/xxxx/Debug-22/root/app/appsettings.json
test/xxxx/Debug-22/root/app/ApolloDeliveryBookingAvailableWatcher.deps.json
test/xxxx/Debug-22/root/app/appsettings.LIVE-1.json
```

Will explore `https://myregistry.company.com` and store the results in `./explore`