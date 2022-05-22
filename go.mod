module fortio.org/proxy

go 1.18

require (
	fortio.org/fortio v1.32.0-test-2
	github.com/fortio/net v0.0.0-20220521234057-cd8eba16ed62 // replace with golang.org/x/net once https://github.com/golang/go/issues/52882 is merged
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
)

require (
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
)
