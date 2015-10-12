Stickify
===================
![Stickify logo](https://raw.githubusercontent.com/ansonl/stickify-pusher/master/stickify-logo-256.png)

View your Microsoft Windows [Sticky Notes](http://windows.microsoft.com/en-us/windows7/using-sticky-notes) anywhere.

###### Code repositories on Github:  [Stickify Pusher](https://github.com/ansonl/stickify-pusher), [Stickify Server](https://github.com/ansonl/stickify-server), [Stickify web app](https://github.com/ansonl/stickify-web-app).

↓ Stickify Server
===================

 1. Clone 
 2. **Pre-built executable is in the `dist` folder if you do not want to compile Stickify Pusher yourself.**  
	 - This executable connects to the *stickify.herokuapp.com* server. 
		 - Provided server is set to wipe nicknames and associated sticky notes if Sticky Pusher has not contacted the server in 24 hours. 
		 - [Stickify Server source](https://github.com/ansonl/stickify-server). 
	 - If you get an error about a missing DLL, your computer is missing the Microsoft Visual C++ Redistributable for VS 2015. The patch can be downloaded [here](http://www.microsoft.com/en-us/download/details.aspx?id=48145). Pyinstaller currently has an issue with bundling the required library in an executable [#1588](https://github.com/pyinstaller/pyinstaller/issues/1588).
 3. You may use the [Stickify web app](https://stickify.gq) to view sticky notes on your phone/other computer. 
	 - [Stickify web app source](https://github.com/ansonl/stickify-web-app). 


Running Stickify Server
-------------

 - Install Go
 - Get Stickify Server code
```
git clone https://github.com/ansonl/stickify-server.git
```
 - Build `stickify-server.go`, `stickify-server` will be compiled.

```
go build stickify-server.go
```
- Set $PORT environment variable to specify port for Stickify Server to listen on. 
```
export PORT=80 #stickify-server will listen on port 80
```
- Run `stickify-server`
```
./stickify-server
```
Stickify Server has only been tested on Linux and OS X. 