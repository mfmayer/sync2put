# sync2put
Synchronizes all files within a directory via http to a server (e.g. by using PUT method to synchronize to a WebDAV server). Besides that other http methods can be used if files shall be synced to any other kind of server.

## Install
sync2put is a written in go and can therefore be installed via: `go get github.com/mfmayer/sync2put` (go must be installed).

## Usage of sync2put
```
$ sync2put --help
Usage of sync2put:
  -append
        Append file name to URL (default true)
  -auth string
        Basic authentication in the form: "<user>:<pwd>"
  -dir string
        Directory to sync (e.g. "/home/mfmayer/dir_to_sync")
  -method string
        HTTP Method to use (default "PUT")
  -s    Synchronize whole directory on start (default true)
  -url string
        Target URL where to sync files to (e.g. "http://192.168.200.1:3001/rsc/")
```