# helloWorldHttp
A simple http server for me to remember the capabilities.
Variables:

* ListeningPort
* Username
* Password
* UsernameQueryParam
* UsernameHeadParam
* PasswordQueryParam
* PasswordHeadParam

can be set on compilation:
```Bash
go build -ldflags "-X main.Username=newUsername -X main.Password=newPassword -X ListeningPort=80 -X main.UsernameQueryParam=u -X main.UsernameHeadParam=u -X main.PasswordQueryParam=p -X main.PasswordHeadParam=p" helloWorld.go
```
or can be passed as arguments:
```Bash
./helloWorld -port 8081 --usernameQueryParam=user --passwordQueryParam pass
```

#### logging http.Handler
Logs every request
 
#### authenticate http.Handler
Checks all Username and Password Parameters for accepted values; Ignores favicon.ico
