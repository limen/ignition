# Ignitor examples

## Configuration

See [.env.yml](https://github.com/limen/ignitor/blob/master/examples/.env.yml)
. Modify it according to your environment.

```
locale: zh_cn
timezone: Asia/Shanghai
dbhost: '127.0.0.1'
dbport: 5432
dbname: ignitor
dbuser: ignitor
dbpassword: goignitor
```

## Create users table

Create table ``ignitor_users`` with 3 columns.
- id - int primary key
- username - varchar(30)
- password - varchar(256)

## Run

```
$ go run main.go
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /user-info                --> main.main.func1 (3 handlers)
[GIN-debug] POST   /users                    --> main.main.func2 (3 handlers)
[GIN-debug] GET    /config                   --> main.main.func3 (3 handlers)
[GIN-debug] GET    /panic                    --> main.main.func4 (3 handlers)
[GIN-debug] Listening and serving HTTP on :8765
```

## Send requests

```
$ curl --data "username=1_&password=123" http://localhost:8765/users
{"code":"ParamError","data":{"password":["password should contains at least 6 characters"],"username":["username should contains 3-30 characters","shouldn't contain '_'"]},"msg":"error","status":"SUCCESS"}
```

```
$ curl --data "username=orange&password=Orange" http://localhost:8765/users
{"code":"SUCCESS","data":{"username":"orange"},"msg":"success","status":"SUCCESS"}
```

```
$ curl http://localhost:8765/user-info?username=orange
{"code":"SUCCESS","data":{"error":null,"user":{"id":4,"username":"orange","password":"Orange"}},"msg":"success","status":"SUCCESS"}
```

```
$ curl http://localhost:8765/config?getters=locale
{"code":"SUCCESS","data":{"locale":"zh_cn"},"msg":"success","status":"SUCCESS"}
```

```
$ curl http://localhost:8765/config?getters=locale,tz
{"code":"SUCCESS","data":{"locale":"zh_cn","tz":"Asia/Shanghai"},"msg":"success","status":"SUCCESS"}
```

## Develop

see [main.go](https://github.com/limen/ignitor/blob/master/examples/main.go)