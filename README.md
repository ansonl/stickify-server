Stickify
===================
![Stickify logo](https://raw.githubusercontent.com/ansonl/stickify-pusher/master/stickify-logo-256.png)

View your Microsoft Windows [Sticky Notes](http://windows.microsoft.com/en-us/windows7/using-sticky-notes) anywhere.

###### Code repositories on Github:  [Stickify Pusher](https://github.com/ansonl/stickify-pusher), [Stickify Server](https://github.com/ansonl/stickify-server), [Stickify web app](https://github.com/ansonl/stickify-web-app).

↓ Stickify Server
===================
 - User account setup is ad hoc. A user's PIN and sticky notes will be wiped if the user does not "send stickies" in a certain amount of time from [Sticky Pusher](https://github.com/ansonl/stickify-pusher/). This expiry period can be adjusted in the `userExpireSeconds` global variable in the code. 
	 - `userExpireSeconds` is currently set to 24 hours.
	 - The expiry for a user is reset when whenever the user updates stickies through a successful POST request to `/update`.

Running Stickify Server
-------------

 - Install [Go](https://golang.org/doc/install/source)
 - Get Stickify Server code
```
git clone https://github.com/ansonl/stickify-server.git
```
 - Build `stickify-server.go`, `stickify-server` will be compiled.

```
go build stickify-server.go
```
- Set `PORT` environment variable to specify port for Stickify Server to listen on. 
```
export PORT=80 #stickify-server will listen on port 80
```
- Run `stickify-server`
```
./stickify-server
```
Notes
-------------
Stickify Server has only been tested on Linux and OS X. 