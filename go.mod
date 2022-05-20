module fortio.org/proxy

go 1.18

require (
	fortio.org/fortio v1.31.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2
)

require (
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect

)

replace golang.org/x/net => github.com/fortio/golang-net v0.0.0-20220519234753-9259fd44fa73 // Has https://go.dev/cl/407454
